package main

import (
	"fmt"
	"log"
	"net"
	"os"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
	"github.com/AkiBatra25/rentrelay/internal/matching"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	port := envOrDefault("GRPC_PORT", "50056")

	propertyConn, err := grpc.NewClient(
		envOrDefault("PROPERTY_SERVICE_ADDR", "localhost:50052"),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("create property service client: %v", err)
	}
	defer propertyConn.Close()

	landlordConn, err := grpc.NewClient(
		envOrDefault("LANDLORD_SERVICE_ADDR", "localhost:50053"),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("create landlord service client: %v", err)
	}
	defer landlordConn.Close()

	agreementConn, err := grpc.NewClient(
		envOrDefault("AGREEMENT_SERVICE_ADDR", "localhost:50055"),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("create agreement service client: %v", err)
	}
	defer agreementConn.Close()

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("listen on port %s: %v", port, err)
	}

	server := grpc.NewServer()
	rentrelaypb.RegisterMatchingServiceServer(server, matching.NewService(
		rentrelaypb.NewPropertyServiceClient(propertyConn),
		rentrelaypb.NewLandlordServiceClient(landlordConn),
		rentrelaypb.NewAgreementServiceClient(agreementConn),
	))

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	reflection.Register(server)

	fmt.Printf("matching-service listening on :%s\n", port)
	if err := server.Serve(listener); err != nil {
		log.Fatalf("serve grpc: %v", err)
	}
}

func envOrDefault(name string, fallback string) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	return value
}
