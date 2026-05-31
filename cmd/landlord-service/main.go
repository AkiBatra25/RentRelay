package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
	"github.com/AkiBatra25/rentrelay/internal/landlord"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50053"
	}

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("listen on port %s: %v", port, err)
	}

	propertyAddr := os.Getenv("PROPERTY_SERVICE_ADDR")
	if propertyAddr == "" {
		propertyAddr = "localhost:50052"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	propertyConn, err := grpc.DialContext(ctx, propertyAddr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Fatalf("connect to property service at %s: %v", propertyAddr, err)
	}
	defer propertyConn.Close()

	propertyClient := rentrelaypb.NewPropertyServiceClient(propertyConn)

	server := grpc.NewServer()
	rentrelaypb.RegisterLandlordServiceServer(server, landlord.NewInMemoryService(propertyClient))

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	reflection.Register(server)

	fmt.Printf("landlord-service listening on :%s\n", port)
	if err := server.Serve(listener); err != nil {
		log.Fatalf("serve grpc: %v", err)
	}
}
