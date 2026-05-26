# RentRelay — Full Architecture Document

## 1. Problem Statement

Rental disputes are one of the most common civil cases in Indian courts.
The root cause is almost always the same: no neutral, auditable, tamper-evident
record of what was agreed, what was paid, when it was paid, and what happened when
something went wrong. Landlords and tenants use WhatsApp threads. When disputes
arise there is no ground truth.

RentRelay is a cloud-native microservice platform that manages the full lifecycle
of a rental relationship: agreement creation and signing → escrow holding of
security deposit → monthly rent payment tracking → dispute filing and resolution
→ agreement termination. Every state transition is recorded with a quorum-written,
replicated audit log.

---

## 2. Microservice Map (PDF Project 1 pattern)

| PDF Service       | RentRelay Equivalent     | Responsibility                                      |
|-------------------|--------------------------|-----------------------------------------------------|
| User Service      | User Service             | Tenant + landlord profiles, JWT auth, RBAC          |
| Driver Service    | Landlord Service         | Property registration, availability, lease terms   |
| Rider Service     | Tenant Service           | Rental requests, lease status, payment history      |
| Matching Service  | Matching Service         | Match tenant requests to available properties       |
| Trip Service      | Agreement Service        | Full agreement + escrow lifecycle state machine      |
| Notification Svc  | Notification Service     | SMS, email, in-app push on every state change       |
| Location Service  | Document Service         | Signed document hashing, storage, verification     |
| Station Service   | Property Service         | Property metadata, nearby amenities, zone mapping   |

---

## 3. Distributed Storage Topology (PDF Project 2 pattern)

```
                        ┌─────────────────────┐
                        │   CONTROLLER NODE   │
                        │                     │
                        │  - Key space map    │
                        │  - Partition table  │
                        │  - Heartbeat recv   │
                        │  - Rebalance logic  │
                        └──────────┬──────────┘
                                   │  gRPC
              ┌────────────────────┼────────────────────┐
              │                    │                    │
    ┌─────────▼──────┐   ┌─────────▼──────┐   ┌────────▼───────┐
    │   WORKER-1     │   │   WORKER-2     │   │   WORKER-3     │
    │                │   │                │   │                │
    │ Agreements A-F │   │ Agreements G-M │   │ Agreements N-S │
    │ Primary shard  │   │ Primary shard  │   │ Primary shard  │
    └────────────────┘   └────────────────┘   └────────────────┘
              │
    ┌─────────▼──────┐
    │   WORKER-4     │
    │                │
    │ Agreements T-Z │
    │ Primary shard  │
    └────────────────┘

Each record is replicated to 2 additional workers → 3 total replicas.
Write quorum = 2 of 3. 3rd replica written asynchronously.
On worker failure: controller detects via missed heartbeat (10s),
reassigns primary responsibility, triggers re-replication from surviving replicas.
```

---

## 4. Escrow State Machine (Agreement Service)

```
                      ┌───────────────┐
                      │    CREATED    │  ← Agreement draft created
                      └──────┬────────┘
                             │ both_parties_sign()
                      ┌──────▼────────┐
                      │    SIGNED     │  ← Digital signatures captured
                      └──────┬────────┘
                             │ tenant_pays_deposit()
                      ┌──────▼────────┐
                      │ ESCROW_HELD   │  ← Security deposit locked
                      └──────┬────────┘
                             │ lease_starts()
                      ┌──────▼────────┐
                      │    ACTIVE     │  ← Rent payments tracked monthly
                      └──┬───────┬────┘
                         │       │
          dispute_raised()│       │notice_period_started()
                         │       │
              ┌──────────▼─┐  ┌──▼──────────────┐
              │  DISPUTED  │  │  NOTICE_PERIOD   │
              └──────┬─────┘  └──────┬───────────┘
                     │               │ notice_completed()
          resolved() │        ┌──────▼────────┐
                     │        │  TERMINATING  │
              ┌──────▼─────┐  └──────┬────────┘
              │  RESOLVED  │         │ property_vacated()
              └────────────┘  ┌──────▼────────┐
                              │  COMPLETED    │  ← Deposit returned/deducted
                              └───────────────┘
```

---

## 5. Data Flow: Monthly Rent Payment

```
Tenant App
    │
    │  POST /payments/rent  (REST Gateway)
    ▼
API Gateway (Go + gRPC client)
    │
    │  gRPC: TenantService.InitiatePayment()
    ▼
Tenant Service
    │
    │  gRPC: AgreementService.RecordPayment()
    ▼
Agreement Service
    │
    ├─── 1. Write to PRIMARY worker (quorum write start)
    ├─── 2. Write to REPLICA-1 worker  ──── await both ACK
    └─── 3. Write to REPLICA-2 worker  ──── async (background)
    │
    │  (2-of-3 quorum reached → return success)
    │
    │  gRPC: NotificationService.Send()
    ▼
Notification Service
    │
    ├─── SMS to landlord: "Rent received for May 2025"
    └─── Push to tenant: "Payment confirmed, receipt #1234"
```

---

## 6. Data Flow: Dispute Filing

```
Landlord App
    │  POST /disputes  (REST Gateway)
    ▼
API Gateway
    │  gRPC: LandlordService.RaiseDispute()
    ▼
Landlord Service
    │  gRPC: AgreementService.TransitionState(DISPUTED)
    ▼
Agreement Service
    │
    ├─── Fetch agreement from controller (get worker for this key)
    ├─── Transition state: ACTIVE → DISPUTED
    ├─── Quorum write new state (2-of-3)
    ├─── gRPC: DocumentService.LockDocuments(agreementId)
    └─── gRPC: NotificationService.Send() → both parties
    │
    │  Dispute record created, escrow frozen
    ▼
Dispute Resolution UI (React)
    Admin reviews → resolves → escrow released/deducted
```

---

## 7. Service Communication Matrix

```
Service              Calls                          Protocol
────────────────     ─────────────────────────────  ──────────
API Gateway          All services                   gRPC + REST
User Service         Notification Service           gRPC
Landlord Service     Property Service               gRPC
                     Matching Service               gRPC
                     Agreement Service              gRPC
Tenant Service       Matching Service               gRPC
                     Agreement Service              gRPC
Matching Service     Landlord Service               gRPC
                     Tenant Service                 gRPC
                     Agreement Service              gRPC
                     Notification Service           gRPC
Agreement Service    Controller (storage)           gRPC
                     Document Service               gRPC
                     Notification Service           gRPC
Notification Svc     (external: SMS/email APIs)     HTTP
Document Service     Controller (storage)           gRPC
Property Service     Controller (storage)           gRPC
Controller           Worker nodes (4)               gRPC
```

---

## 8. Kubernetes Deployment Summary

| Component           | Kind          | Replicas    | HPA Min/Max  |
|---------------------|---------------|-------------|--------------|
| API Gateway         | Deployment    | 2           | 2 / 5        |
| User Service        | Deployment    | 2           | 2 / 4        |
| Landlord Service    | Deployment    | 2           | 2 / 4        |
| Tenant Service      | Deployment    | 2           | 2 / 4        |
| Matching Service    | Deployment    | 1           | 1 / 5  ← PDF |
| Agreement Service   | Deployment    | 2           | 2 / 4        |
| Notification Svc    | Deployment    | 2           | 2 / 5        |
| Document Service    | Deployment    | 2           | 2 / 3        |
| Property Service    | Deployment    | 2           | 2 / 3        |
| Storage Controller  | StatefulSet   | 1           | —            |
| Storage Worker      | StatefulSet   | 4           | —            |
| MongoDB             | StatefulSet   | 3 (replica) | —            |
| Redis               | StatefulSet   | 1           | —            |
| Prometheus          | Deployment    | 1           | —            |
| Grafana             | Deployment    | 1           | —            |

---

## 9. Failure Scenarios & Recovery

### Matching Service Pod Crash
- K8s liveness probe fails → pod restarted automatically
- HPA scales from 1 → up to 5 pods under load
- Pending match requests in Redis queue consumed by new pods
- Demo: `kubectl delete pod matching-service-xxx` → watch recovery

### Storage Worker Node Failure
- Worker stops sending heartbeat to Controller (10s timeout)
- Controller marks worker DEAD
- Controller identifies all primary shards on dead worker
- Controller promotes replica on surviving worker to primary
- Controller initiates re-replication to restore 3-replica count
- Clients querying controller get updated routing table
- Demo: `kubectl delete pod storage-worker-2` → watch rebalance

### Quorum Write Failure (only 1 replica available)
- Agreement Service gets ACK from only 1 of 3 workers
- Returns 503 to caller with `QUORUM_NOT_MET` error
- Client retries with exponential backoff
- No partial write committed (rollback on surviving replicas)

---

## 10. Tech Stack Summary

| Layer            | Technology                          |
|------------------|-------------------------------------|
| Services         | Go (all 8 microservices)            |
| Inter-service    | gRPC (Protocol Buffers v3)          |
| REST Gateway     | Go + chi router                     |
| Primary DB       | MongoDB 7 (replica set, 3 nodes)    |
| Storage layer    | Custom Go controller + 4 workers    |
| Cache / Locks    | Redis 7 (Redlock for escrow locks)  |
| Messaging        | Kafka (notification fanout)         |
| Orchestration    | Kubernetes (K8s 1.29)               |
| Package mgmt     | Helm 3                              |
| Monitoring       | Prometheus + Grafana                |
| Auth             | JWT (RS256) + RBAC middleware       |
| CI/CD            | GitHub Actions → Docker → Helm      |
