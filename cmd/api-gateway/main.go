package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
	"github.com/AkiBatra25/rentrelay/cmd/api-gateway/handlers"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Read service addresses from env, fall back to localhost defaults
	userAddr      := envOrDefault("USER_SERVICE_ADDR",      "localhost:50051")
	propertyAddr  := envOrDefault("PROPERTY_SERVICE_ADDR",  "localhost:50052")
	agreementAddr := envOrDefault("AGREEMENT_SERVICE_ADDR", "localhost:50055")
	httpPort      := envOrDefault("PORT",                   "8080")

	// Connect to each gRPC service
	// insecure means no TLS — fine for local dev
	userConn,      err := grpc.NewClient(userAddr,      grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil { log.Fatalf("connect user service: %v", err) }
	defer userConn.Close()

	propertyConn,  err := grpc.NewClient(propertyAddr,  grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil { log.Fatalf("connect property service: %v", err) }
	defer propertyConn.Close()

	agreementConn, err := grpc.NewClient(agreementAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil { log.Fatalf("connect agreement service: %v", err) }
	defer agreementConn.Close()

	// Create gRPC clients from the connections
	userClient      := rentrelaypb.NewUserServiceClient(userConn)
	propertyClient  := rentrelaypb.NewPropertyServiceClient(propertyConn)
	agreementClient := rentrelaypb.NewAgreementServiceClient(agreementConn)

	// Create handlers — each handler wraps one gRPC client
	userH      := handlers.NewUserHandler(userClient)
	propertyH  := handlers.NewPropertyHandler(propertyClient)
	agreementH := handlers.NewAgreementHandler(agreementClient)

	// Register HTTP routes
	mux := http.NewServeMux()

	// Health check — test this first!
	mux.HandleFunc("/health", handlers.HealthHandler)

	// User routes
	mux.HandleFunc("/api/users/register", userH.Register)
	mux.HandleFunc("/api/users/login",    userH.Login)
	mux.HandleFunc("/api/users/",         userH.GetUser) // GET /api/users/{id}

	// Property routes
	mux.HandleFunc("/api/properties/search", propertyH.SearchProperties)
	mux.HandleFunc("/api/properties/",       func(w http.ResponseWriter, r *http.Request) {
		// Route to register or get depending on method
		if r.Method == http.MethodPost {
			propertyH.RegisterProperty(w, r)
		} else {
			propertyH.GetProperty(w, r)
		}
	})

	// Agreement routes
	mux.HandleFunc("/api/agreements/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if r.Method == http.MethodPost && path == "/api/agreements/" {
			agreementH.CreateAgreement(w, r)
			return
		}
		// /api/agreements/{id}/sign
		if r.Method == http.MethodPost {
			agreementH.SignAgreement(w, r)
			return
		}
		agreementH.GetAgreement(w, r)
	})
	mux.HandleFunc("/api/agreements", agreementH.CreateAgreement)

	fmt.Printf("api-gateway listening on :%s\n", httpPort)
	fmt.Println("  GET  /health")
	fmt.Println("  POST /api/users/register")
	fmt.Println("  POST /api/users/login")
	fmt.Println("  GET  /api/users/{id}")
	fmt.Println("  POST /api/properties/")
	fmt.Println("  GET  /api/properties/search?city=Bengaluru&max_rent=30000&bedrooms=2")
	fmt.Println("  GET  /api/properties/{id}")
	fmt.Println("  POST /api/agreements")
	fmt.Println("  GET  /api/agreements/{id}")
	fmt.Println("  POST /api/agreements/{id}/sign")

	log.Fatal(http.ListenAndServe(":"+httpPort, mux))
}

func envOrDefault(name, fallback string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return fallback
}
