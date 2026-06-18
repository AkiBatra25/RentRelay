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

### Planned

- Agreement Service integration with replicated storage
- Kubernetes deployment validation
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
│   └── 03-monitoring.yaml
│
├── scripts/
│   └── generate-proto.ps1
│
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

- Go
- Git
- Docker Desktop

Verify:

```powershell
go version
git --version
docker version
docker compose version
```

---

## Setup

Clone the repository:

```powershell
git clone <your-repository-url>
cd RentRelay\rentrelay
```

Install Go dependencies:

```powershell
go mod tidy
```

Run tests:

```powershell
go test -buildvcs=false ./...
```

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

```powershell
$env:MONGO_URI="mongodb://rentrelay:rentrelay@localhost:27017/rentrelay?authSource=admin"
$env:MONGO_DATABASE="rentrelay"
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

---

## MongoDB

Start MongoDB locally:

```powershell
cd C:\IIITB\Academics\RentRelay\rentrelay
docker compose up -d mongodb
```

Check status:

```powershell
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

```powershell
docker compose up -d mongodb
```

Start UserService:

```powershell
cd C:\IIITB\Academics\RentRelay\rentrelay
$env:MONGO_URI="mongodb://rentrelay:rentrelay@localhost:27017/rentrelay?authSource=admin"
$env:MONGO_DATABASE="rentrelay"
go run -buildvcs=false ./cmd/user-service
```

In another terminal:

```powershell
cd C:\IIITB\Academics\RentRelay\rentrelay
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

```powershell
docker compose up -d mongodb
```

Start PropertyService:

```powershell
cd C:\IIITB\Academics\RentRelay\rentrelay
$env:MONGO_URI="mongodb://rentrelay:rentrelay@localhost:27017/rentrelay?authSource=admin"
$env:MONGO_DATABASE="rentrelay"
go run -buildvcs=false ./cmd/property-service
```

In another terminal:

```powershell
cd C:\IIITB\Academics\RentRelay\rentrelay
go run -buildvcs=false ./cmd/property-smoke
```

Expected output:

```text
registered property_id=property-...
search results=1
updated availability=false
```

---

## Run API Gateway

Start MongoDB and all backend services first:

```bash
docker compose up -d mongodb
bash start-all.sh
```

In a new terminal, start the gateway:

```bash
cd "/mnt/d/e drive/RentRelay"
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

## Run Tests

From the Go module root:

```powershell
cd C:\IIITB\Academics\RentRelay\rentrelay
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

```powershell
cd C:\IIITB\Academics\RentRelay\rentrelay
docker build -f Dockerfile.user-service -t rentrelay/user-service:local .
```

Run MongoDB:

```powershell
docker compose up -d mongodb
```

Run UserService container:

```powershell
docker run --rm `
  --name rentrelay-user-service `
  --network rentrelay_default `
  -e MONGO_URI="mongodb://rentrelay:rentrelay@mongodb:27017/rentrelay?authSource=admin" `
  -e MONGO_DATABASE="rentrelay" `
  -p 50051:50051 `
  rentrelay/user-service:local
```

Run smoke client:

```powershell
go run -buildvcs=false ./cmd/user-smoke
```

---

## Docker: PropertyService

Build the image:

```powershell
cd C:\IIITB\Academics\RentRelay\rentrelay
docker build -f Dockerfile.property-service -t rentrelay/property-service:local .
```

Run MongoDB:

```powershell
docker compose up -d mongodb
```

Run PropertyService container:

```powershell
docker run --rm `
  --name rentrelay-property-service `
  --network rentrelay_default `
  -e MONGO_URI="mongodb://rentrelay:rentrelay@mongodb:27017/rentrelay?authSource=admin" `
  -e MONGO_DATABASE="rentrelay" `
  -p 50052:50052 `
  rentrelay/property-service:local
```

Run smoke client:

```powershell
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

```powershell
docker compose up --build
```

Stop containers:

```powershell
docker compose down
```

Stop containers and remove volumes:

```powershell
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

### `go test ./...` says module not found

Run Go commands from:

```powershell
C:\IIITB\Academics\RentRelay\rentrelay
```

because this is where `go.mod` exists.

The outer folder is the Git repository root.

### Port already in use

Example:

```text
listen tcp :50052: bind: Only one usage of each socket address is normally permitted
```

Check the process:

```powershell
netstat -ano | findstr :50052
```

Kill it:

```powershell
taskkill /PID <PID> /F
```

### MongoDB connection refused

Start MongoDB first:

```powershell
docker compose up -d mongodb
docker compose ps
```

Then start the service.

### `localhost` vs `mongodb`

From your host machine:

```text
localhost:27017
```

From another Docker container:

```text
mongodb:27017
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

```powershell
gofmt -w internal/property/service.go
go test -buildvcs=false ./...
git status
git add .
git commit -m "meaningful message"
git push
```

---

## Architecture Goal

The final system is planned as:

```text
Client
  |
  | REST/JSON
  v
API Gateway
  |
  | gRPC
  v
Microservices
  |
  | MongoDB / custom storage / async messaging
  v
Data and infrastructure layer
```

Planned service map:

```text
User Service
Property Service
Landlord Service
Tenant Service
Matching Service
Agreement Service
Notification Service
Document Service
Storage Controller
Storage Workers
API Gateway (implemented)
```

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
```

---

## Resume Summary

RentRelay is a cloud-native rental agreement platform built with Go, gRPC, Protocol Buffers, MongoDB, Docker, and Kubernetes.

Current implemented milestone:

```text
Implemented ten Go-based gRPC microservices for users, properties, landlords, tenants, matching, agreements, notifications, documents, and distributed storage, with protobuf contracts, MongoDB persistence, Dockerized infrastructure, service-to-service gRPC calls, replicated quorum writes, write-ahead log crash recovery, watchdog-based failure detection, a REST API gateway for HTTP/JSON clients, and full smoke-test validation.
```

Possible resume bullet:

```text
Built a cloud-native backend in Go using gRPC, Protocol Buffers, MongoDB, Docker, and Kubernetes manifests, implementing ten microservice foundations plus partitioned distributed storage with primary-replica routing, 2-of-3 quorum writes, write-ahead log crash recovery, automatic dead worker detection via heartbeat watchdog, and a REST API gateway exposing HTTP/JSON endpoints for all core workflows.
```

---

## License

This project is currently for academic and learning purposes.