#!/bin/bash

set -e  # Avsluta direkt vid fel

LOG_FILE="experiment_log.txt"
DURATION_NORMAL=60
DURATION_ATTACK=60
DURATION_RECOVERY=30

echo "======================================="
echo "  AUTO SETUP + RUN EXPERIMENT SCRIPT"
echo "======================================="

# 0. Kontrollera dependencies
check_command() {
    if ! command -v "$1" &> /dev/null; then
        echo "[ERROR] $1 is not installed. Please install it first."
        exit 1
    fi
}

echo "[*] Checking dependencies..."
check_command docker
check_command kubectl
check_command minikube
check_command python3

# 1. Starta Minikube
echo "[*] Starting Minikube (if not running)..."
minikube status &> /dev/null || minikube start

echo "[*] Pointing Docker to Minikube..."
eval $(minikube docker-env)

# 2. Bygg External DB
echo "[*] Building ExternalDB image..."
docker build -t external-db:local ./externalDB

echo "[*] Starting External DB container..."
docker rm -f my-external-db 2>/dev/null || true
docker run -d --name my-external-db -p 3306:3306 external-db:local

# 3. Bygg applikationer
echo "[*] Building application images..."
docker build -t app1:local ./apps/app1
docker build -t app2:local ./apps/app2
docker build -t app3:local ./apps/app3
docker build -t sidecar:local ./sidecar

# 4. Deploy till Kubernetes
echo "[*] Applying Kubernetes manifests..."
kubectl apply -f apps/app3/external-service.yaml
kubectl apply -f apps/app3/app3-deployment.yaml
kubectl apply -f apps/app2/app2-deployment.yaml
kubectl apply -f apps/app1/app1-deployment.yaml
kubectl apply -f sidecar/sidecar-deployment.yaml

# 5. Vänta på pods
echo "[*] Waiting for pods to be ready..."
kubectl wait --for=condition=ready pod --all --timeout=180s

# 6. Kör experiment
echo "--- Environment ready, starting experiment ---"
> "$LOG_FILE"

echo "[$(date +%T)] FAS1_START" >> "$LOG_FILE"
echo ">>> Phase 1: Normal ($DURATION_NORMAL s)..."

kubectl logs -f --timestamps deployment/sidecar-verifier >> "$LOG_FILE" &
LOG_PID=$!

sleep $DURATION_NORMAL

echo "[$(date +%T)] FAS2_START" >> "$LOG_FILE"
echo ">>> Phase 2: Killing DB ($DURATION_ATTACK s)..."

docker stop my-external-db
sleep $DURATION_ATTACK

echo "[$(date +%T)] FAS3_START" >> "$LOG_FILE"
echo ">>> Phase 3: Recovery ($DURATION_RECOVERY s)..."

docker start my-external-db
sleep $DURATION_RECOVERY

# Stoppa loggning
kill $LOG_PID 2>/dev/null || true

echo "[$(date +%T)] EXPERIMENT_END" >> "$LOG_FILE"

# 7. Analys
# echo "[*] Running analysis..."
python3 sidecar/analyze.py "$LOG_FILE"

# 8. Cleanup (VIKTIGT)
echo "[*] Cleaning up environment..."

kubectl delete -f sidecar/sidecar-deployment.yaml --ignore-not-found
kubectl delete -f apps/app1/app1-deployment.yaml --ignore-not-found
kubectl delete -f apps/app2/app2-deployment.yaml --ignore-not-found
kubectl delete -f apps/app3/app3-deployment.yaml --ignore-not-found
kubectl delete -f apps/app3/external-service.yaml --ignore-not-found

docker rm -f my-external-db 2>/dev/null || true

echo "======================================="
echo "  DONE: Experiment completed cleanly"
echo "======================================="
