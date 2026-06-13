package main

import (
	"context"
	"fmt"
	"log"
	"time"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	controllerConn := mustConnect("localhost:50060")
	defer controllerConn.Close()
	controller := rentrelaypb.NewStorageControllerClient(controllerConn)

	key := fmt.Sprintf("agreement-storage-%d", time.Now().UnixNano())
	route, err := controller.GetWorkerForKey(ctx, &rentrelaypb.GetWorkerRequest{Key: key})
	if err != nil {
		log.Fatalf("route key: %v", err)
	}

	targets := append([]*rentrelaypb.PartitionInfo{route.Primary}, route.Replicas...)
	acks := 0
	for index, target := range targets {
		conn := mustConnect(hostWorkerAddress(target))
		client := rentrelaypb.NewStorageWorkerClient(conn)
		resp, err := client.Put(ctx, &rentrelaypb.KVPutRequest{
			Key: key, Value: []byte("quorum-protected-agreement"), ReplicaVersion: 1, IsPrimary: index == 0,
		})
		conn.Close()
		if err == nil && resp.Success {
			acks++
		}
	}
	if acks < 2 {
		log.Fatalf("quorum failed: acknowledgements=%d", acks)
	}

	conn := mustConnect(hostWorkerAddress(route.Primary))
	defer conn.Close()
	value, err := rentrelaypb.NewStorageWorkerClient(conn).Get(ctx, &rentrelaypb.KVGetRequest{Key: key})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("key=%s primary=%s replicas=%d quorum_acks=%d found=%v value=%s\n",
		key, route.Primary.WorkerId, len(route.Replicas), acks, value.Found, string(value.Value))
}

func hostWorkerAddress(worker *rentrelaypb.PartitionInfo) string {
	switch worker.WorkerId {
	case "storage-worker-0":
		return "localhost:51061"
	case "storage-worker-1":
		return "localhost:51062"
	case "storage-worker-2":
		return "localhost:51063"
	case "storage-worker-3":
		return "localhost:51064"
	default:
		return worker.WorkerAddress
	}
}

func mustConnect(address string) *grpc.ClientConn {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("connect %s: %v", address, err)
	}
	return conn
}
