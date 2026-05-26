package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	addr := os.Getenv("USER_SERVICE_ADDR")
	if addr == "" {
		addr = "localhost:50051"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("create grpc client: %v", err)
	}
	defer conn.Close()

	client := rentrelaypb.NewUserServiceClient(conn)
	email := fmt.Sprintf("smoke-%d@test.com", time.Now().UnixNano())

	registerResp, err := client.Register(ctx, &rentrelaypb.RegisterRequest{
		Name:     "Smoke Test User",
		Email:    email,
		Phone:    "9999999999",
		Password: "pass123",
		Role:     rentrelaypb.UserRole_ROLE_TENANT,
	})
	if err != nil {
		log.Fatalf("register user: %v", err)
	}

	loginResp, err := client.Login(ctx, &rentrelaypb.LoginRequest{
		Email:    email,
		Password: "pass123",
	})
	if err != nil {
		log.Fatalf("login user: %v", err)
	}

	fmt.Printf("registered user_id=%s email=%s\n", registerResp.User.UserId, registerResp.User.Email)
	fmt.Printf("login token prefix=%s\n", loginResp.Token[:9])
}
