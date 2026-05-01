#!/bin/bash

set -e

LOG_FILE="experiment_log.txt"
DURATION_NORMAL=60
DURATION_ATTACK=60
DURATION_RECOVERY=30

echo "======================================="
echo "   SETUP + RUN EXPERIMENT (UBUNTU)"
echo "======================================="

# =====================================
# 0. Install dependencies (Ubuntu)
# =====================================

install_if_missing() {
    CMD=$1
    PKG=$2

    if ! command -v $CMD &> /dev/null; then
        echo "[!] $CMD not found. Installing $PKG..."
        sudo apt update
        sudo apt install -y $PKG
    else
        echo "[OK] $CMD already installed"
    fi
}

echo "[*] Checking dependencies..."

install_if_missing docker docker.io
install_if_missing kubectl kubectl
install_if_missing python3 python3
install_if_missing curl curl

# 1. Docker setup
echo "[*] Ensuring Docker is running..."
sudo systemctl enable docker
sudo systemctl start docker

if ! groups $USER | grep -q docker; then
    echo "[*] Adding user to docker group..."
    sudo usermod -aG docker $USER
    echo "[!] Log out and back in, then run this script again."
    exit 1
fi

# 2. Install Minikube
if ! command -v minikube &> /dev/null; then
    echo "[!] Installing Minikube..."

    curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
    sudo install minikube-linux-amd64 /usr/local/bin/minikube
    rm minikube-linux-amd64

    echo "[OK] Minikube installed"
else
    echo "[OK] Minikube already installed"
fi

# 3. Start Minikube
echo "[*] Starting Minikube..."

if ! minikube status &> /dev/null; then
    minikube start --driver=docker
else
    echo "[OK] Minikube already running"
fi

echo "[*] Pointing Docker to Minikube..."
eval $(minikube docker-env)

# 4. Build External DB
echo "[*] Building ExternalDB image..."
docker build -t external-db:local ./externalDB

echo "[*] Starting External DB container..."
docker rm -f my-external-db 2>/dev/null || true
docker run -d --name my-external-db -p 3306:3306 external-db:local

# 5. Build applications
echo "[*] Building application images..."
docker build -t app1:local ./apps/app1
docker build -t app2:local ./apps/app2
docker build -t app3:local ./apps/app3
docker build -t sidecar:local ./sidecar

# 6. Deploy to Kubernetes
echo "[*] Applying Kubernetes manifests..."
kubectl apply -f apps/app3/external-service.yaml
kubectl apply -f apps/app3/app3-deployment.yaml
kubectl apply -f apps/app2/app2-deployment.yaml
kubectl apply -f apps/app1/app1-deployment.yaml
kubectl apply -f sidecar/sidecar-deployment.yaml

# 7. Wait for pods
echo "[*] Waiting for pods to be ready..."
kubectl wait --for=condition=ready pod --all --timeout=180s

# 8. Run experiment
echo "--- Environment ready, starting experiment ---"
> "$LOG_FILE"

echo "[$(date +%T)] FAS1_START" >> "$LOG_FILE"
echo ">>> Phase 1: Normal ($DURATION_NORMAL s)..."

kubectl logs -f --timestamps deployment/sidecar-verifier >> "$LOG_FILE" &
LOG_PID=$!

sleep $DURATION_NORMAL

echo "[$(date +%T)] FAS2_START" >> "$LOG_FILE"
echo ">>> Phase 2: Stopping DB ($DURATION_ATTACK s)..."

docker stop my-external-db
sleep $DURATION_ATTACK

echo "[$(date +%T)] FAS3_START" >> "$LOG_FILE"
echo ">>> Phase 3: Recovery ($DURATION_RECOVERY s)..."

docker start my-external-db
sleep $DURATION_RECOVERY

kill $LOG_PID 2>/dev/null || true
echo "[$(date +%T)] EXPERIMENT_END" >> "$LOG_FILE"

# 9. Analyze results
echo "[*] Running analysis..."
python3 sidecar/analyze.py "$LOG_FILE"

# 10. Cleanup
echo "[*] Cleaning up..."

kubectl delete -f sidecar/sidecar-deployment.yaml --ignore-not-found
kubectl delete -f apps/app1/app1-deployment.yaml --ignore-not-found
kubectl delete -f apps/app2/app2-deployment.yaml --ignore-not-found
kubectl delete -f apps/app3/app3-deployment.yaml --ignore-not-found
kubectl delete -f apps/app3/external-service.yaml --ignore-not-found

docker rm -f my-external-db 2>/dev/null || true

echo "======================================="
echo "   DONE: Experiment completed"
echo "======================================="
