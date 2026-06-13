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
	addr := os.Getenv("TENANT_SERVICE_ADDR")
	if addr == "" {
		addr = "localhost:50054"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("create grpc client: %v", err)
	}
	defer conn.Close()

	client := rentrelaypb.NewTenantServiceClient(conn)
	tenantID := fmt.Sprintf("tenant-smoke-%d", time.Now().UnixNano())

	created, err := client.CreateRentalRequest(ctx, &rentrelaypb.CreateRentalRequestReq{
		TenantId:       tenantID,
		PreferredCity:  "Bengaluru",
		PreferredZone:  "south",
		BedroomsNeeded: 2,
		MaxRent:        30000,
		Furnishing:     rentrelaypb.FurnishingType_SEMI_FURNISHED,
	})
	if err != nil {
		log.Fatalf("create rental request: %v", err)
	}

	fetched, err := client.GetRentalRequest(ctx, &rentrelaypb.GetUserRequest{UserId: tenantID})
	if err != nil {
		log.Fatalf("get rental request: %v", err)
	}

	dashboard, err := client.GetDashboard(ctx, &rentrelaypb.GetUserRequest{UserId: tenantID})
	if err != nil {
		log.Fatalf("get dashboard: %v", err)
	}

	fmt.Printf("created rental_request_id=%s tenant_id=%s\n", created.RequestId, created.TenantId)
	fmt.Printf("fetched rental_request city=%s max_rent=%.2f\n", fetched.PreferredCity, fetched.MaxRent)
	fmt.Printf("dashboard tenant_id=%s months_paid=%d\n", dashboard.TenantId, dashboard.MonthsPaid)
}
