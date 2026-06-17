#!/bin/bash
export MONGO_URI="mongodb://rentrelay:rentrelay@localhost:27017/rentrelay?authSource=admin"
export MONGO_DATABASE="rentrelay"
export WORKER_ID="worker-1"
export WORKER_ADDRESS="localhost:50061"

echo "Starting storage worker on :50061..."
go run -buildvcs=false ./cmd/storage-worker &
sleep 5

echo "Starting storage controller on :50060..."
go run -buildvcs=false ./cmd/storage-controller &

echo "Storage started! Waiting 10 seconds..."
sleep 10
echo "Ready!"
wait
