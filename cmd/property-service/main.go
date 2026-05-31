package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
	"github.com/AkiBatra25/rentrelay/internal/property"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50052"
	}

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("listen on port %s: %v", port, err)
	}

	server := grpc.NewServer()
	propertyService, closeStore := newPropertyService()
	defer closeStore()

	rentrelaypb.RegisterPropertyServiceServer(server, propertyService)

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	reflection.Register(server)

	fmt.Printf("property-service listening on :%s\n", port)
	if err := server.Serve(listener); err != nil {
		log.Fatalf("serve grpc: %v", err)
	}
}

func newPropertyService() (*property.Service, func()) {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Println("MONGO_URI not set; using in-memory property repository")
		return property.NewInMemoryService(), func() {}
	}

	database := os.Getenv("MONGO_DATABASE")
	if database == "" {
		database = "rentrelay"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	repo, err := property.NewMongoRepository(ctx, mongoURI, database)
	if err != nil {
		log.Fatalf("connect to MongoDB: %v", err)
	}

	log.Printf("connected property-service to MongoDB database %q", database)
	return property.NewService(repo), func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := repo.Close(ctx); err != nil {
			log.Printf("close MongoDB connection: %v", err)
		}
	}
}
