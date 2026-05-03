#!/bin/bash

LOG_FILE="experiment_log.txt"
DURATION_NORMAL=60
DURATION_ATTACK=60
DURATION_RECOVERY=30
CLUSTER_NAME="projektkurs"

echo "--- Startar kind-kluster ---"

if ! kind get clusters | grep -q "$CLUSTER_NAME"; then
    kind create cluster --name "$CLUSTER_NAME"
else
    echo "[*] Kluster '$CLUSTER_NAME' finns redan, återanvänder det."
fi

ACTUAL_NETWORK=$(docker network ls --filter "name=kind" --format "{{.Name}}" | head -n 1)
echo "[*] Använder Docker-nätverk: $ACTUAL_NETWORK"

echo "[*] Building ExternalDB image..."
docker build -t external-db:local ./externalDB

echo "[*] Starting External DB container..."
docker stop my-external-db 2>/dev/null || true
docker rm my-external-db 2>/dev/null || true
docker run -d \
    --name my-external-db \
    --network "$ACTUAL_NETWORK" \
    -p 3306:3306 \
    external-db:local

echo "[*] Väntar på att databasen ska starta..."
sleep 15

echo "[*] Building and loading application images..."

docker build -t app1:local ./apps/app1
docker save app1:local | kind load image-archive /dev/stdin --name "$CLUSTER_NAME"

docker build -t app2:local ./apps/app2
docker save app2:local | kind load image-archive /dev/stdin --name "$CLUSTER_NAME"

docker build -t app3:local ./apps/app3
docker save app3:local | kind load image-archive /dev/stdin --name "$CLUSTER_NAME"

docker build -t sidecar:local ./sidecar
docker save sidecar:local | kind load image-archive /dev/stdin --name "$CLUSTER_NAME"

echo "[*] Applying Kubernetes manifests..."
kubectl apply -f apps/app3/external-service.yaml
kubectl apply -f apps/app3/app3-deployment.yaml
kubectl apply -f apps/app2/app2-deployment.yaml
kubectl apply -f apps/app1/app1-deployment.yaml
kubectl apply -f sidecar/sidecar-deployment.yaml

echo "[*] Väntar på att pods ska starta..."
kubectl wait --for=condition=ready pod --all --timeout=120s

echo "[*] Startar om pods for att nollstalla raknare..."
kubectl rollout restart deployment/app3-deployment
kubectl rollout restart deployment/sidecar-verifier
kubectl wait --for=condition=ready pod --all --timeout=60s
sleep 5

echo "--- Miljon ar igang, startar experiment ---"
> "$LOG_FILE"

echo "[$(date -u +%T)] FAS1_START" >> "$LOG_FILE"
echo "[*] Fas 1: Normalt lage ($DURATION_NORMAL sekunder)..."
kubectl logs -f --timestamps deployment/sidecar-verifier >> "$LOG_FILE" &
LOG_PID=$!
sleep $DURATION_NORMAL

echo "[$(date -u +%T)] FAS2_START" >> "$LOG_FILE"
echo "[*] Fas 2: Stanger av databasen ($DURATION_ATTACK sekunder)..."
docker stop my-external-db
sleep $DURATION_ATTACK

echo "[$(date -u +%T)] FAS3_START" >> "$LOG_FILE"
echo "[*] Fas 3: Startar databasen igen ($DURATION_RECOVERY sekunder)..."
docker start my-external-db
sleep $DURATION_RECOVERY

kill $LOG_PID 2>/dev/null
echo "[$(date -u +%T)] EXPERIMENT_SLUT" >> "$LOG_FILE"

echo "--- Experiment klart, kor python-analys ---"
python3 sidecar/analyze.py "$LOG_FILE"
