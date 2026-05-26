package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
	"github.com/AkiBatra25/rentrelay/internal/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50051"
	}

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("listen on port %s: %v", port, err)
	}

	server := grpc.NewServer()
	userService, closeStore := newUserService()
	defer closeStore()
	rentrelaypb.RegisterUserServiceServer(server, userService)

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	reflection.Register(server)

	fmt.Printf("user-service listening on :%s\n", port)
	if err := server.Serve(listener); err != nil {
		log.Fatalf("serve grpc: %v", err)
	}
}

func newUserService() (*user.Service, func()) {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Println("MONGO_URI not set; using in-memory user repository")
		return user.NewInMemoryService(), func() {}
	}

	database := os.Getenv("MONGO_DATABASE")
	if database == "" {
		database = "rentrelay"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	repo, err := user.NewMongoRepository(ctx, mongoURI, database)
	if err != nil {
		log.Fatalf("connect to MongoDB: %v", err)
	}

	log.Printf("connected user-service to MongoDB database %q", database)
	return user.NewService(repo), func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := repo.Close(ctx); err != nil {
			log.Printf("close MongoDB connection: %v", err)
		}
	}
}
