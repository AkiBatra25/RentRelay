#!/bin/bash

set -e

echo "===== Starting Minikube ====="
minikube start --driver=docker

echo "===== Using Minikube Docker ====="
eval $(minikube docker-env)

echo "===== Building Docker Images ====="

for service in \
user-service \
property-service \
landlord-service \
tenant-service \
agreement-service \
matching-service \
notification-service \
document-service \
storage-controller \
storage-worker \
api-gateway
do
    echo "Building $service..."
    docker build \
        -f Dockerfile.$service \
        -t rentrelay/$service:latest .
done

echo "===== Applying Kubernetes Configuration ====="

kubectl apply -f k8s/00-namespace-config.yaml

echo ""
echo "===== Checking MongoDB Secret ====="

if kubectl get secret rentrelay-secrets -n rentrelay >/dev/null 2>&1; then
    echo "MongoDB secret already exists."
else
    echo "ERROR: MongoDB secret does not exist."
    echo ""
    echo "Create it first using:"
    echo 'read -s -p "Paste MongoDB Atlas URI: " MONGO_URI'
    echo 'kubectl create secret generic rentrelay-secrets --namespace=rentrelay --from-literal=MONGO_URI="$MONGO_URI"'
    echo 'unset MONGO_URI'
    exit 1
fi

echo ""
echo "===== Deploying Services ====="

kubectl apply -f k8s/user-service.yaml
kubectl apply -f k8s/property-service.yaml
kubectl apply -f k8s/tenant-service.yaml
kubectl apply -f k8s/landlord-service.yaml

kubectl apply -f k8s/storage-controller.yaml

kubectl apply -f k8s/storage-worker-0.yaml
kubectl apply -f k8s/storage-worker-1.yaml
kubectl apply -f k8s/storage-worker-2.yaml
kubectl apply -f k8s/storage-worker-3.yaml

kubectl apply -f k8s/agreement-service.yaml
kubectl apply -f k8s/notification-service.yaml
kubectl apply -f k8s/document-service.yaml
kubectl apply -f k8s/matching-service.yaml
kubectl apply -f k8s/api-gateway.yaml

echo ""
echo "===== Waiting for Pods ====="

kubectl wait \
    --for=condition=Ready \
    pods \
    --all \
    -n rentrelay \
    --timeout=300s

echo ""
echo "===== Deployment Status ====="

kubectl get pods -n rentrelay

echo ""
echo "======================================"
echo "RentRelay deployed successfully!"
echo "======================================"
echo ""
echo "To access the API, run:"
echo ""
echo "kubectl port-forward -n rentrelay service/api-gateway 8080:8080"
echo ""
echo "Then open:"
echo ""
echo "http://localhost:8080/health"
