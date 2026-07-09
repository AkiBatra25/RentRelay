# RentRelay

RentRelay is a cloud-native rental agreement and escrow management platform built as a learning-focused microservices project using Go, gRPC, Protocol Buffers, MongoDB, Docker, Docker Compose, and Kubernetes.

The project models a real distributed backend system for rental workflows such as user registration, property listing, tenant-property matching, rental agreements, escrow tracking, rent payments, disputes, documents, and notifications.

---

## Current Status

RentRelay currently has ten implemented backend service foundations:

### Implemented

- Go module setup
- Protocol Buffers API contract
- Generated Go gRPC code
- User Service
  - Register user
  - Login user
  - Get user
  - Update KYC
  - Refresh token
  - In-memory repository
  - MongoDB repository
  - gRPC smoke client
  - Docker support
- Property Service
  - Register property
  - Get property
  - Search properties
  - Update availability
  - List properties by landlord
  - In-memory repository
  - MongoDB repository
  - gRPC smoke client
- Landlord Service
  - Set lease terms
  - Get lease terms
  - Get landlord dashboard
  - Calls Property Service over gRPC
  - In-memory repository
  - MongoDB repository
  - gRPC smoke client
- Tenant Service
  - Create rental request
  - Get rental request
  - Get tenant dashboard placeholder
  - In-memory repository
  - MongoDB repository
  - gRPC smoke client
- Matching Service
  - Search available properties through Property Service
  - Fetch optional lease terms through Landlord Service
  - Score and rank match candidates
  - gRPC smoke client
  - Docker support
- Agreement Service
  - Agreement creation and retrieval
  - Two-party signing
  - Escrow hold
  - Lease start and termination lifecycle
  - MongoDB repository
  - State-machine validation tests
  - Replicated storage integration via ReplicatedRepository
  - Every state change written to MongoDB and storage workers simultaneously
  - FindByID falls back to storage layer if MongoDB is unavailable
  - Requires 2-of-3 quorum acknowledgement before confirming write
  - Enabled by setting STORAGE_CONTROLLER_ADDR environment variable
- Notification Service
  - Send and broadcast notifications
  - Notification history
  - Live in-process subscriber streams
  - MongoDB persistence with TTL index
- Document Service
  - Upload and SHA-256 hashing
  - Hash verification
  - Agreement document listing
  - Document locking and unlocking
  - MongoDB persistence
- Storage Controller
  - Registers storage workers and records heartbeats
  - Partitions keys across a 256-slot hash space
  - Selects one primary and two replica workers
  - Exposes partition-table and routing RPCs
  - Watchdog goroutine detects dead workers after 3 missed heartbeats
  - Automatically marks workers as unavailable and stops routing to them
- Storage Worker
  - Stores versioned key-value records in memory
  - Supports put, get, delete, and key listing
  - Supports client-streamed key transfer
  - Participates in replicated 2-of-3 quorum writes
  - Write-ahead log persists every put and delete to disk before applying to memory
  - Replays WAL on startup to restore full state after crash or restart
  - Each worker maintains its own WAL file at a configurable path
- Local MongoDB using Docker Compose
- Docker Compose integration for implemented services
- Kubernetes manifests drafted for the larger system
- Dockerfiles for all 11 services
  - Multi-stage builds using golang:1.26-alpine builder and alpine:3.20 runner
  - Produces minimal images with only the compiled binary
  - All services containerized and tested locally
- API Gateway
  - REST HTTP/JSON gateway on port 8080
  - POST /api/users/register — register a new user
  - POST /api/users/login — login and get a token
  - GET  /api/users/{id} — get user by ID
  - POST /api/properties — register a new property
  - GET  /api/properties/search?city=&max_rent=&bedrooms= — search properties
  - GET  /api/properties/{id} — get property by ID
  - POST /api/agreements — create a rental agreement
  - GET  /api/agreements/{id} — get agreement by ID
  - POST /api/agreements/{id}/sign — sign an agreement
  - GET  /health — health check endpoint
- GitHub Actions CI/CD pipeline
  - Runs all tests on every push to main
  - Builds all 11 Docker images if tests pass
  - Pushes images to Docker Hub automatically
  - No manual build or push steps needed
- Kubernetes deployment on Minikube (validated locally)
  - All 15 pods running simultaneously including 4 storage workers
  - Namespace, ConfigMap, and Secrets configured
  - Individual k8s manifests for every service
  - API Gateway exposed as NodePort service
  - Agreement replication confirmed active through storage-controller
  - WAL replay confirmed on storage worker startup
  - All HTTP endpoints tested and verified through kubectl port-forward
  - deploy-local.sh script automates the full deployment

### Planned

- Cloud deployment on AWS, GCP, or Azure

---

## Why This Project Exists

Rental workflows often involve scattered agreements, informal payment tracking, unclear deposit handling, and weak dispute records.

RentRelay is designed to explore how a real cloud-native backend could manage:

- User registration and identity
- Property listing
- Tenant-property matching
- Agreement lifecycle
- Security deposit escrow
- Payment receipts
- Dispute workflows
- Document verification
- Notifications
- Distributed storage concepts

The project is intentionally built step by step to learn cloud computing, backend development, distributed systems, and deployment practices.

---

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go |
| Internal communication | gRPC |
| API contract | Protocol Buffers |
| Database | MongoDB |
| Local infrastructure | Docker Compose |
| Containerization | Docker |
| Orchestration target | Kubernetes |
| Monitoring draft | Prometheus and Grafana |
| Architecture style | Microservices |

---

## Project Structure

```text
rentrelay/
├── cmd/
│   ├── user-service/
│   │   └── main.go
│   ├── user-smoke/
│   │   └── main.go
│   ├── property-service/
│   │   └── main.go
│   └── property-smoke/
│       └── main.go
│
├── gen/go/
│   ├── rentrelay.pb.go
│   └── rentrelay_grpc.pb.go
│
├── internal/
│   ├── user/
│   │   ├── repository.go
│   │   ├── memory_repository.go
│   │   ├── mongo_repository.go
│   │   ├── service.go
│   │   └── service_test.go
│   │
│   └── property/
│       ├── repository.go
│       ├── memory_repository.go
│       ├── memory_repository_test.go
│       ├── mongo_repository.go
│       ├── service.go
│       └── service_test.go
│   └── storageworker/
│       ├── service.go
│       ├── service_test.go
│       └── wal.go
│   └── storagecontroller/
│       └── service.go
│
├── proto/
│   └── rentrelay.proto
│
├── mongo/
│   └── schemas.js
│
├── k8s/
│   ├── 00-namespace-config.yaml
│   ├── 01-services.yaml
│   ├── 02-storage-cluster.yaml
│   ├── 03-monitoring.yaml
│   ├── user-service.yaml
│   ├── property-service.yaml
│   ├── landlord-service.yaml
│   ├── tenant-service.yaml
│   ├── agreement-service.yaml
│   ├── matching-service.yaml
│   ├── notification-service.yaml
│   ├── document-service.yaml
│   ├── storage-controller.yaml
│   ├── storage-worker-0.yaml
│   ├── storage-worker-1.yaml
│   ├── storage-worker-2.yaml
│   ├── storage-worker-3.yaml
│   └── api-gateway.yaml
│
├── scripts/
│   └── generate-proto.ps1
│
├── start-all.sh
├── test-all.sh
├── deploy-local.sh
├── compose.yaml
├── Dockerfile.user-service
├── Dockerfile.property-service
├── .env.example
├── go.mod
└── go.sum
```

---

## Core Concepts Learned So Far

### Protocol Buffers

The file:

```text
proto/rentrelay.proto
```

defines the official service contract.

It describes:

- what services exist
- what RPC methods they expose
- what request data they accept
- what response data they return

Example:

```proto
service UserService {
  rpc Register(RegisterRequest) returns (RegisterResponse);
  rpc Login(LoginRequest) returns (LoginResponse);
}
```

This means:

```text
UserService has a Register method and a Login method.
Each method accepts a specific request type and returns a specific response type.
```

---

### gRPC

gRPC is the communication system used between services.

Instead of calling normal HTTP JSON endpoints, services call strongly typed RPC methods generated from the `.proto` file.

In RentRelay:

```text
UserService listens on port 50051
PropertyService listens on port 50052
```

A gRPC client can call:

```text
Register
Login
RegisterProperty
SearchProperties
```

as if they were normal functions, even though the call goes over the network.

---

### Generated Code

The generated files are:

```text
gen/go/rentrelay.pb.go
gen/go/rentrelay_grpc.pb.go
```

These are created from:

```text
proto/rentrelay.proto
```

They contain:

- Go structs for protobuf messages
- Go interfaces for gRPC services
- client code
- server registration code

These files should not be edited manually.

Regenerate them when the proto file changes.

---

### Microservice Layout

Each service follows this pattern:

```text
cmd/<service-name>/main.go
```

starts the server.

```text
internal/<domain>/service.go
```

contains business logic.

```text
internal/<domain>/repository.go
```

defines the storage interface.

```text
internal/<domain>/memory_repository.go
```

stores data in memory for tests and simple local runs.

```text
internal/<domain>/mongo_repository.go
```

stores data in MongoDB.

This pattern is currently used for:

```text
UserService
PropertyService
```

---

### Repository Pattern

The repository pattern separates business logic from storage logic.

For example, `PropertyService` depends on:

```go
repo Repository
```

not directly on MongoDB.

That means the same service can use:

```go
NewService(NewMemoryRepository())
```

or:

```go
NewService(NewMongoRepository(...))
```

without rewriting the service methods.

This makes the code easier to test, change, and maintain.

---

### Docker

Docker packages an application into a container image.

A Docker image contains:

- the compiled service binary
- the runtime environment
- the startup command
- exposed ports

This makes the service portable across machines and cloud environments.

---

### Docker Compose

Docker Compose runs multiple containers together.

RentRelay currently uses Compose for MongoDB and service development.

For example:

```text
MongoDB container
UserService container
PropertyService container
```

can run together on the same local Docker network.

Inside Docker Compose, services can talk by service name.

For example:

```text
mongodb:27017
```

From the host machine, MongoDB is reached using:

```text
localhost:27017
```

---

## Implemented Services

## UserService

### Purpose

UserService handles basic user account workflows.

### Current Features

- Register a user
- Login a user
- Get user by ID
- Update KYC status
- Refresh token
- Store users in memory
- Store users in MongoDB
- Run as a gRPC server
- Validate using a smoke client

### Port

```text
50051
```

### Main Files

```text
cmd/user-service/main.go
cmd/user-smoke/main.go
internal/user/service.go
internal/user/repository.go
internal/user/memory_repository.go
internal/user/mongo_repository.go
```

### Supported RPCs

```text
Register
Login
GetUser
UpdateKYC
RefreshToken
```

---

## PropertyService

### Purpose

PropertyService handles rental property listing and search workflows.

### Current Features

- Register a property
- Get property by ID
- Search available properties
- Update property availability
- List properties by landlord
- Store properties in memory
- Store properties in MongoDB
- Run as a gRPC server
- Validate using a smoke client

### Port

```text
50052
```

### Main Files

```text
cmd/property-service/main.go
cmd/property-smoke/main.go
internal/property/service.go
internal/property/repository.go
internal/property/memory_repository.go
internal/property/mongo_repository.go
```

### Supported RPCs

```text
RegisterProperty
GetProperty
SearchProperties
UpdateAvailability
ListByLandlord
```

---

## Prerequisites

Install:

- Go 1.26+ from https://go.dev/dl
- Git from https://git-scm.com
- Docker Desktop from https://www.docker.com/products/docker-desktop

Verify:

```bash
go version
git --version
docker version
docker compose version
```

---

## Setup

### Clone and install

```bash
git clone https://github.com/AkiBatra25/RentRelay
cd RentRelay
go mod tidy
```

### Run all tests

```bash
go test -buildvcs=false ./...
```

All 10 internal packages should show ok.

### Start MongoDB

```bash
docker compose up -d mongodb
docker compose ps
```

Wait until status shows healthy.

### Start all services at once

```bash
bash start-all.sh
```

Wait for the message: ALL SERVICES RUNNING

### Run all smoke tests

Open a new terminal:

```bash
bash test-all.sh
```

All 9 services should show PASSED.

### Start the API gateway

Open another terminal:

```bash
go run -buildvcs=false ./cmd/api-gateway
```

Test it:

```bash
curl http://localhost:8080/health
```

Expected: {"status":"ok","time":"..."}

---

## Regenerate Protobuf Code

Whenever `proto/rentrelay.proto` changes, regenerate Go code:

```powershell
cd C:\IIITB\Academics\RentRelay\rentrelay
.\scripts\generate-proto.ps1
```

Then run:

```powershell
go test -buildvcs=false ./...
```

---

## Environment Variables

The services use environment variables for configuration.

Example:

```bash
export MONGO_URI="mongodb://rentrelay:rentrelay@localhost:27017/rentrelay?authSource=admin"
export MONGO_DATABASE="rentrelay"
```

Common variables:

| Variable | Purpose |
|---|---|
| `MONGO_URI` | MongoDB connection string |
| `MONGO_DATABASE` | MongoDB database name |
| `GRPC_PORT` | gRPC server port |
| `USER_SERVICE_ADDR` | Address used by UserService smoke client |
| `PROPERTY_SERVICE_ADDR` | Address used by PropertyService smoke client |
| `WORKER_ID` | Unique ID for a storage worker (e.g. worker-1) |
| `WORKER_ADDRESS` | Address the worker advertises to the controller (e.g. localhost:50061) |
| `SHARD_START` | Start of hash slot range this worker owns (0-255) |
| `SHARD_END` | End of hash slot range this worker owns (0-255) |
| `WAL_PATH` | Path to write-ahead log file (default: /tmp/<worker-id>.log) |
| `CONTROLLER_ADDR` | Address of storage controller (default: localhost:50060) |
| `HTTP_PORT` | Port for the REST API gateway (default: 8080) |
| `AGREEMENT_SERVICE_ADDR` | Address of agreement service (default: localhost:50055) |
| `STORAGE_CONTROLLER_ADDR` | If set, Agreement Service replicates to storage workers |

---

## MongoDB

Start MongoDB locally:

```bash
docker compose up -d mongodb
```

Check status:

```bash
docker compose ps
```

Expected:

```text
rentrelay-mongodb   Up ... healthy
```

Host connection string:

```text
mongodb://rentrelay:rentrelay@localhost:27017/rentrelay?authSource=admin
```

Container-to-container connection string:

```text
mongodb://rentrelay:rentrelay@mongodb:27017/rentrelay?authSource=admin
```

---

## Run UserService With Go

Start MongoDB:

```bash
docker compose up -d mongodb
```

Start UserService:

```bash
export MONGO_URI="mongodb://rentrelay:rentrelay@localhost:27017/rentrelay?authSource=admin"
export MONGO_DATABASE="rentrelay"
go run -buildvcs=false ./cmd/user-service
```

In another terminal:

```bash
go run -buildvcs=false ./cmd/user-smoke
```

Expected output:

```text
registered user_id=user-...
login token prefix=dev-token
```

---

## Run PropertyService With Go

Start MongoDB:

```bash
docker compose up -d mongodb
```

Start PropertyService:

```bash
export MONGO_URI="mongodb://rentrelay:rentrelay@localhost:27017/rentrelay?authSource=admin"
export MONGO_DATABASE="rentrelay"
go run -buildvcs=false ./cmd/property-service
```

In another terminal:

```bash
go run -buildvcs=false ./cmd/property-smoke
```

Expected output:

```text
registered property_id=property-...
search results=1
updated availability=false
```

---

## Docker Hub

All service images are published automatically to Docker Hub via GitHub Actions.

Pull any image:

```bash
docker pull 30301207/rentrelay-user-service:latest
docker pull 30301207/rentrelay-api-gateway:latest
```

Available images:

| Image | Port |
|---|---|
| 30301207/rentrelay-api-gateway | 8080 |
| 30301207/rentrelay-user-service | 50051 |
| 30301207/rentrelay-property-service | 50052 |
| 30301207/rentrelay-landlord-service | 50053 |
| 30301207/rentrelay-tenant-service | 50054 |
| 30301207/rentrelay-agreement-service | 50055 |
| 30301207/rentrelay-matching-service | 50056 |
| 30301207/rentrelay-notification-service | 50057 |
| 30301207/rentrelay-document-service | 50058 |
| 30301207/rentrelay-storage-controller | 50060 |
| 30301207/rentrelay-storage-worker | 50061-50063 |

---

## CI/CD Pipeline

Every push to the main branch triggers the GitHub Actions pipeline:

```text
Push to main
    ↓
Run all tests (go test ./...)
    ↓ (only if tests pass)
Build 11 Docker images
    ↓
Push to Docker Hub
```

To add secrets for the pipeline to work on a fork:

Go to Settings → Secrets and variables → Actions and add:

| Secret | Value |
|---|---|
| DOCKER_USERNAME | your Docker Hub username |
| DOCKER_PASSWORD | your Docker Hub password |

---

## Run API Gateway

Start MongoDB and all backend services first:

```bash
docker compose up -d mongodb
bash start-all.sh
```

In a new terminal, start the gateway:

```bash
go run -buildvcs=false ./cmd/api-gateway
```

Test endpoints:

```bash
# Health check
curl http://localhost:8080/health

# Register a user
curl -s -X POST http://localhost:8080/api/users/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Harshita","email":"h@test.com","phone":"9999999999","password":"pass123","role":"tenant"}'

# Search properties
curl -s "http://localhost:8080/api/properties/search?city=Bengaluru&max_rent=30000&bedrooms=2"

# Create an agreement
curl -s -X POST http://localhost:8080/api/agreements \
  -H "Content-Type: application/json" \
  -d '{"tenant_id":"t-1","landlord_id":"l-1","property_id":"p-1","monthly_rent":25000,"deposit_amount":75000,"lease_months":11,"notice_days":30}'
```

---

## Kubernetes Deployment (Minikube)

This deploys the full RentRelay stack on a local Kubernetes cluster using Minikube.

### Prerequisites

Install Minikube:

```bash
curl -LO https://github.com/kubernetes/minikube/releases/latest/download/minikube-linux-amd64
sudo install minikube-linux-amd64 /usr/local/bin/minikube
minikube version
```

kubectl comes with Docker Desktop. Verify:

```bash
kubectl version --client
```

### Deploy

```bash
# Start Minikube
minikube start --driver=docker

# Point Docker to Minikube internal registry
eval $(minikube docker-env)

# Create namespace and config
kubectl apply -f k8s/00-namespace-config.yaml

# Create MongoDB secret (paste your MongoDB URI when prompted)
read -s -p "Paste MongoDB URI: " MONGO_URI
kubectl create secret generic rentrelay-secrets \
  --namespace=rentrelay \
  --from-literal=MONGO_URI="$MONGO_URI"
unset MONGO_URI

# Run the full automated deploy script
bash deploy-local.sh
```

### Check all pods are running

```bash
kubectl get pods -n rentrelay
```

Expected — all 15 pods showing Running:

```text
agreement-service       Running
api-gateway             Running
document-service        Running
landlord-service        Running
matching-service        Running
notification-service    Running
property-service        Running
storage-controller      Running
storage-worker-0        Running
storage-worker-1        Running
storage-worker-2        Running
storage-worker-3        Running
tenant-service          Running
user-service            Running
```

### Access the API

```bash
kubectl port-forward -n rentrelay service/api-gateway 8080:8080
```

In another terminal:

```bash
curl http://localhost:8080/health
```

### Check logs

```bash
kubectl logs -n rentrelay deployment/user-service
kubectl logs -n rentrelay deployment/agreement-service
kubectl logs -n rentrelay deployment/storage-worker-0
```

### Stop Minikube

```bash
minikube stop
```

### Delete everything

```bash
minikube delete
```

---

## Run Tests

From the Go module root:

```bash
go test -buildvcs=false ./...
```

Expected output includes:

```text
ok github.com/AkiBatra25/rentrelay/internal/user
ok github.com/AkiBatra25/rentrelay/internal/property
```

Some packages may show:

```text
[no test files]
```

That is normal for command packages such as:

```text
cmd/user-service
cmd/property-service
```

---

## Docker: UserService

Build the image:

```bash
docker build -f Dockerfile.user-service -t rentrelay/user-service:local .
```

Run MongoDB:

```bash
docker compose up -d mongodb
```

Run UserService container:

```bash
docker run --rm \
  --name rentrelay-user-service \
  --network rentrelay_default \
  -e MONGO_URI="mongodb://rentrelay:rentrelay@mongodb:27017/rentrelay?authSource=admin" \
  -e MONGO_DATABASE="rentrelay" \
  -p 50051:50051 \
  rentrelay/user-service:local
```

Run smoke client:

```bash
go run -buildvcs=false ./cmd/user-smoke
```

---

## Docker: PropertyService

Build the image:

```bash
docker build -f Dockerfile.property-service -t rentrelay/property-service:local .
```

Run MongoDB:

```bash
docker compose up -d mongodb
```

Run PropertyService container:

```bash
docker run --rm \
  --name rentrelay-property-service \
  --network rentrelay_default \
  -e MONGO_URI="mongodb://rentrelay:rentrelay@mongodb:27017/rentrelay?authSource=admin" \
  -e MONGO_DATABASE="rentrelay" \
  -p 50052:50052 \
  rentrelay/property-service:local
```

Run smoke client:

```bash
go run -buildvcs=false ./cmd/property-smoke
```

---

## Docker Compose

MongoDB is defined in:

```text
compose.yaml
```

The goal is for Compose to run the local system:

```text
MongoDB
UserService
PropertyService
```

Typical command:

```bash
docker compose up --build
```

Stop containers:

```bash
docker compose down
```

Stop containers and remove volumes:

```bash
docker compose down -v
```

Only use `-v` when you are okay deleting local MongoDB data.

---

## Important Ports

| Component | Port |
|---|---|
| API Gateway (HTTP) | 8080 |
| UserService | 50051 |
| PropertyService | 50052 |
| LandlordService | 50053 |
| TenantService | 50054 |
| AgreementService | 50055 |
| MatchingService | 50056 |
| NotificationService | 50057 |
| DocumentService | 50058 |
| Storage Controller | 50060 |
| Storage Worker 1 | 50061 |
| Storage Worker 2 | 50062 |
| Storage Worker 3 | 50063 |
| MongoDB | 27017 |

---

## Common Issues

### Port already in use

```bash
# Check which process is using the port
lsof -t -i:50051

# Kill it
kill $(lsof -t -i:50051)
```

### MongoDB connection refused

Start MongoDB first:

```bash
docker compose up -d mongodb
docker compose ps
```

Then start the service.

### localhost vs mongodb

From your host machine:

```text
localhost:27017
```

From another Docker container:

```text
mongodb:27017
```

### Minikube pods stuck in Pending

Check available resources:

```bash
minikube status
kubectl describe pod <pod-name> -n rentrelay
```

Restart Minikube with more memory:

```bash
minikube delete
minikube start --driver=docker --memory=4096 --cpus=2
```

---

## Development Workflow

Usual workflow:

```text
1. Understand the proto contract
2. Update proto if needed
3. Regenerate generated Go code
4. Implement service logic
5. Implement repository interface
6. Add memory repository
7. Add MongoDB repository
8. Add tests
9. Add smoke client
10. Add Dockerfile
11. Add Docker Compose entry
12. Run tests
13. Run smoke client
14. Commit and push
```

Useful commands:

```bash
gofmt -w internal/property/service.go
go test -buildvcs=false ./...
git status
git add .
git commit -m "meaningful message"
git push
```

---

## Architecture

```text
Client (curl / Postman / browser)
  |
  | HTTP/JSON
  v
API Gateway (:8080)
  |
  | gRPC
  v
┌─────────────────────────────────────────────┐
│  User Service          (:50051)             │
│  Property Service      (:50052)             │
│  Landlord Service      (:50053)             │
│  Tenant Service        (:50054)             │
│  Agreement Service     (:50055)             │
│  Matching Service      (:50056)             │
│  Notification Service  (:50057)             │
│  Document Service      (:50058)             │
│  Storage Controller    (:50060)             │
│  Storage Workers x4    (:50061-50063)       │
└─────────────────────────────────────────────┘
  |
  | MongoDB driver / gRPC / WAL
  v
MongoDB + Distributed Storage (quorum writes)
```

All services communicate internally via gRPC. External clients use HTTP/JSON through the API Gateway which translates to gRPC.

---

## Learning Progress

Completed learning milestones:

```text
1. Initialized Git project
2. Set up Go module
3. Generated protobuf code
4. Built first gRPC service
5. Added in-memory storage
6. Added MongoDB persistence
7. Added Dockerized local MongoDB
8. Added smoke client for end-to-end validation
9. Dockerized UserService
10. Built PropertyService using same architecture pattern
11. Built all ten gRPC service foundations with smoke clients
12. Implemented distributed storage with quorum writes and heartbeats
13. Added write-ahead log to storage worker for crash recovery
14. Added watchdog to storage controller for automatic dead worker detection
15. Built REST API gateway translating HTTP/JSON to gRPC for all core services
16. Integrated Agreement Service with distributed storage replication and quorum writes
17. Wrote Dockerfiles for all 11 services using multi-stage builds
18. Set up GitHub Actions CI/CD pipeline that tests, builds, and pushes on every commit
19. Deployed all 15 pods on Kubernetes using Minikube with full validation
20. Verified live HTTP API calls through REST gateway backed by gRPC microservices on Kubernetes
```

---

## Resume Summary

RentRelay is a cloud-native rental agreement platform built with Go, gRPC, Protocol Buffers, MongoDB, Docker, and Kubernetes.

Current implemented milestone:

```text
Implemented eleven Go-based gRPC microservices and a REST API gateway for a cloud-native rental platform, with protobuf contracts, MongoDB persistence, distributed storage with WAL crash recovery and 2-of-3 quorum writes, dead worker watchdog, agreement replication with dual writes and disaster recovery fallback, multi-stage Docker builds for all services, a GitHub Actions CI/CD pipeline that automatically tests and publishes 11 Docker images on every commit, and full Kubernetes deployment validated on Minikube with 15 running pods and live API verification.
```

Possible resume bullet:

```text
Built a cloud-native rental backend in Go with gRPC, Protocol Buffers, MongoDB, Docker, Kubernetes, and GitHub Actions CI/CD, implementing eleven microservices, partitioned distributed storage with quorum writes, WAL crash recovery, heartbeat watchdog, agreement replication with automatic fallback, a REST API gateway, a fully automated pipeline that tests and publishes 11 Docker images on every push, and a validated Kubernetes deployment with 15 pods verified via live HTTP API calls.
```

---

## Built By

Akshat Batra — https://github.com/AkiBatra25

Harshita Bansal — https://github.com/Harshita30-bansal

---

## License

This project is currently for academic and learning purposes.