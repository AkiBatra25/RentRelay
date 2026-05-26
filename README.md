# RentRelay — Distributed Rental Agreement & Escrow Service

> Microservice platform for tamper-evident rental agreements, escrow management,
> and dispute resolution. Built with Go + gRPC + MongoDB on Kubernetes.

---

## Repository Structure

```
rentrelay/
├── docs/
│   └── ARCHITECTURE.md        ← Full system design, state machine, data flows
├── proto/
│   └── rentrelay.proto        ← All 10 gRPC service definitions
├── mongo/
│   └── schemas.js             ← All MongoDB collections, indexes, validators
├── k8s/
│   ├── 00-namespace-config.yaml   ← Namespace, ConfigMap, Secrets
│   ├── 01-services.yaml           ← All 8 microservices + API gateway + MongoDB + Redis
│   ├── 02-storage-cluster.yaml    ← Controller + 4 Workers StatefulSets (Proj 2)
│   └── 03-monitoring.yaml         ← Prometheus, Grafana, Ingress, NetworkPolicy
└── scripts/
    └── demo.sh                ← 6 live demo scenarios
```

---

## Quick Deploy

```bash
# 1. Apply all configs
kubectl apply -f k8s/00-namespace-config.yaml
kubectl apply -f k8s/01-services.yaml
kubectl apply -f k8s/02-storage-cluster.yaml
kubectl apply -f k8s/03-monitoring.yaml

# 2. Verify all pods running
kubectl get pods -n rentrelay

# 3. Run full demo
chmod +x scripts/demo.sh
./scripts/demo.sh a
```

---

## Local Development

Start MongoDB:

```powershell
docker compose up -d mongodb
```

Run tests:

```powershell
go test -buildvcs=false ./...
```

Run the User Service with MongoDB persistence:

```powershell
$env:MONGO_URI="mongodb://rentrelay:rentrelay@localhost:27017/rentrelay?authSource=admin"
$env:MONGO_DATABASE="rentrelay"
go run -buildvcs=false ./cmd/user-service
```

In another terminal, run the smoke client:

```powershell
go run -buildvcs=false ./cmd/user-smoke
```

---

## PDF Requirements Checklist

| Requirement                                    | Implementation                                        |
|------------------------------------------------|-------------------------------------------------------|
| 8 microservices                                | User, Landlord, Tenant, Matching, Agreement,          |
|                                                | Notification, Document, Property services             |
| gRPC APIs                                      | rentrelay.proto — 10 services, 50+ RPCs               |
| Matching scales 1 → 5 (HPA)                   | matching-service-hpa in 02-storage-cluster.yaml       |
| Service failure demo                           | demo.sh scenario 2 — delete pods, watch recovery      |
| Controller + 4 workers (Proj 2)                | storage-controller + storage-worker StatefulSets      |
| GET + PUT operations                           | StorageWorker.Get + StorageWorker.Put RPCs             |
| Key partitioned across N workers               | Consistent hash ring, 256 slots / 4 workers           |
| Client queries controller first                | AgreementService calls StorageController.GetWorkerForKey |
| 3 replicas per key                             | Quorum write to 2 + async 3rd replica                 |
| PUT succeeds on 2-of-3 replicas                | QUORUM_REPLICAS=2 in ConfigMap                        |
| Heartbeat from each worker                     | 5s interval, 10s timeout, triggers rebalance          |
| On failure: re-replication from other nodes    | demo.sh scenario 4 — delete worker, watch rebalance   |
| MongoDB                                        | 10 collections, geospatial indexes, TTL indexes       |
| Kubernetes deployment                          | All 4 YAML files, Helm-ready                          |
| Prometheus + Grafana                           | Full scrape config + 6 alert rules                    |

---

## gRPC Port Map

| Service              | gRPC Port |
|----------------------|-----------|
| user-service         | 50051     |
| property-service     | 50052     |
| landlord-service     | 50053     |
| tenant-service       | 50054     |
| agreement-service    | 50055     |
| matching-service     | 50056     |
| notification-service | 50057     |
| document-service     | 50058     |
| storage-controller   | 50060     |
| storage-workers      | 50061     |
