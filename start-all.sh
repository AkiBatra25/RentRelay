#!/bin/bash

export MONGO_URI="mongodb://rentrelay:rentrelay@localhost:27017/rentrelay?authSource=admin"
export MONGO_DATABASE="rentrelay"

echo "================================================"
echo "  RentRelay - Starting all services"
echo "================================================"

# Step 1 - Storage Controller first (workers need to register into it)
echo ""
echo "[1/4] Starting Storage Controller..."
go run -buildvcs=false ./cmd/storage-controller &
sleep 6
echo "  Storage Controller ready on :50060"

# Step 2 - Start 3 storage workers
echo ""
echo "[2/4] Starting 3 Storage Workers..."

WORKER_ID="worker-1" WORKER_ADDRESS="localhost:50061" SHARD_START="0" SHARD_END="99" GRPC_PORT="50061" \
  go run -buildvcs=false ./cmd/storage-worker &

WORKER_ID="worker-2" WORKER_ADDRESS="localhost:50062" SHARD_START="0" SHARD_END="99" GRPC_PORT="50062" \
  go run -buildvcs=false ./cmd/storage-worker &

WORKER_ID="worker-3" WORKER_ADDRESS="localhost:50063" SHARD_START="0" SHARD_END="99" GRPC_PORT="50063" \
  go run -buildvcs=false ./cmd/storage-worker &

sleep 6
echo "  Worker-1 ready on :50061"
echo "  Worker-2 ready on :50062"
echo "  Worker-3 ready on :50063"

# Step 3 - All other services
echo ""
echo "[3/4] Starting all other services..."
go run -buildvcs=false ./cmd/user-service &
go run -buildvcs=false ./cmd/property-service &
go run -buildvcs=false ./cmd/landlord-service &
go run -buildvcs=false ./cmd/tenant-service &
go run -buildvcs=false ./cmd/matching-service &
go run -buildvcs=false ./cmd/agreement-service &
go run -buildvcs=false ./cmd/notification-service &
go run -buildvcs=false ./cmd/document-service &

sleep 10
echo "  All services started!"

echo ""
echo "================================================"
echo "  ALL SERVICES RUNNING! Open a new terminal"
echo "  and run: bash test-all.sh"
echo "================================================"
echo ""
echo "  Press Ctrl+C to stop everything"
echo ""

wait
