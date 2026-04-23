#!/bin/bash


echo "--- Starting Industrial IDS Lab Enviroment ---"

# 1. Prepare Minikube / K8s Docker enviroment
# This ensures that images built are available inside the cluster.

if command -v minikube &> /dev/null; then
	echo "[*] Pointing Docker to Minikube..."
	eval $(minikube docker-env)
fi

# 2. Build the external DB image
echo "[*] Building ExternalDB image..."
cd externalDB
docker build -t external-db:local .
cd ..

# 3. Start the DB in Docker (Outside K8s)
# First, stop and remove any old container to avoid name conflicts
echo "[*] Starting External DB container..."
docker stop my-external-db 2>/dev/null || true
docker rm my-external-db 2>/dev/null || true
docker run -d --name my-external-db -p 3306:3306 external-db:local

# 4. Build application images
docker build -t app1:local ./apps/app1
docker build -t app2:local ./apps/app2
docker build -t app3:local ./apps/app3
docker build -t sidecar:local ./sidecar

# 5. Deploy to kubernetes
echo "[*] Applying Kubernetes manifests..."
kubectl apply -f apps/app3/external-service.yaml
kubectl apply -f apps/app3/app3-deployment.yaml
kubectl apply -f apps/app2/app2-deployment.yaml
kubectl apply -f apps/app1/app1-deployment.yaml
kubectl apply -f sidecar/sidecar-deployment.yaml

echo "--- Enviroment is Booting up! ---"
echo "To ceck the logs, use: kubectl logs -f deployment/sidecar-verifier"
