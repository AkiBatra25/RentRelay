package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
	"github.com/AkiBatra25/rentrelay/internal/storageworker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	workerID := requiredEnv("WORKER_ID")
	workerAddress := requiredEnv("WORKER_ADDRESS")
	shardStart := int32Env("SHARD_START")
	shardEnd := int32Env("SHARD_END")
	port := envOrDefault("GRPC_PORT", "50061")

	// WAL file path — defaults to /tmp/<workerID>.log
	// Each worker gets its own log file so they don't overwrite each other
	walPath := envOrDefault("WAL_PATH", "/tmp/"+workerID+".log")

	controllerConn, err := grpc.NewClient(
		envOrDefault("CONTROLLER_ADDR", "localhost:50060"),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer controllerConn.Close()
	controller := rentrelaypb.NewStorageControllerClient(controllerConn)

	// Use WAL-backed service so data survives restarts
	service, err := storageworker.NewServiceWithWAL(workerID, walPath)
	if err != nil {
		log.Fatalf("open WAL at %s: %v", walPath, err)
	}
	defer service.Close()

	registerWorker(controller, workerID, workerAddress, shardStart, shardEnd)
	go heartbeatLoop(controller, service, workerID)

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}
	server := grpc.NewServer()
	rentrelaypb.RegisterStorageWorkerServer(server, service)
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	reflection.Register(server)
	fmt.Printf("storage-worker %s listening on :%s for shards %d-%d (WAL: %s)\n",
		workerID, port, shardStart, shardEnd, walPath)
	log.Fatal(server.Serve(listener))
}

func registerWorker(client rentrelaypb.StorageControllerClient, id, address string, start, end int32) {
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, err := client.RegisterWorker(ctx, &rentrelaypb.PartitionInfo{
			WorkerId: id, WorkerAddress: address, ShardStart: start, ShardEnd: end, IsAlive: true,
		})
		cancel()
		if err == nil {
			return
		}
		log.Printf("register worker: %v; retrying", err)
		time.Sleep(2 * time.Second)
	}
}

func heartbeatLoop(client rentrelaypb.StorageControllerClient, service *storageworker.Service, workerID string) {
	interval := time.Duration(intEnv("HEARTBEAT_INTERVAL_SEC", 5)) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		_, err := client.Heartbeat(ctx, &rentrelaypb.HeartbeatRequest{WorkerId: workerID, StoredKeys: service.StoredKeys()})
		cancel()
		if err != nil {
			log.Printf("heartbeat: %v", err)
		}
	}
}

func requiredEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatalf("%s is required", name)
	}
	return value
}

func int32Env(name string) int32 {
	value, err := strconv.Atoi(requiredEnv(name))
	if err != nil {
		log.Fatalf("%s must be an integer", name)
	}
	return int32(value)
}

func intEnv(name string, fallback int) int {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
