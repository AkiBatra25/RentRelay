package main

import (
	"fmt"
	"log"
	"net"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
	"github.com/AkiBatra25/rentrelay/internal/storagecontroller"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	service := storagecontroller.NewService()

	// Start the watchdog — it runs in background forever
	// checking worker heartbeats and marking dead workers
	service.StartWatchdog()

	listener, err := net.Listen("tcp", ":50060")
	if err != nil {
		log.Fatal(err)
	}
	server := grpc.NewServer()
	rentrelaypb.RegisterStorageControllerServer(server, service)
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	reflection.Register(server)
	fmt.Println("storage-controller listening on :50060 (watchdog active)")
	log.Fatal(server.Serve(listener))
}
