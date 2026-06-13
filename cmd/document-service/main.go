package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
	"github.com/AkiBatra25/rentrelay/internal/document"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	service, closeStore := newService()
	defer closeStore()
	port := envOrDefault("GRPC_PORT", "50058")
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	server := grpc.NewServer()
	rentrelaypb.RegisterDocumentServiceServer(server, service)
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	reflection.Register(server)
	fmt.Printf("document-service listening on :%s\n", port)
	log.Fatal(server.Serve(listener))
}

func newService() (*document.Service, func()) {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		repo := document.NewMemoryRepository()
		return document.NewService(repo), func() {}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	repo, err := document.NewMongoRepository(ctx, uri, envOrDefault("MONGO_DATABASE", "rentrelay"))
	if err != nil {
		log.Fatalf("connect MongoDB: %v", err)
	}
	return document.NewService(repo), func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = repo.Close(ctx)
	}
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
