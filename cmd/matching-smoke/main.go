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
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	propertyConn := mustConnect(envOrDefault("PROPERTY_SERVICE_ADDR", "localhost:50052"))
	defer propertyConn.Close()
	tenantConn := mustConnect(envOrDefault("TENANT_SERVICE_ADDR", "localhost:50054"))
	defer tenantConn.Close()
	matchingConn := mustConnect(envOrDefault("MATCHING_SERVICE_ADDR", "localhost:50056"))
	defer matchingConn.Close()

	propertyClient := rentrelaypb.NewPropertyServiceClient(propertyConn)
	tenantClient := rentrelaypb.NewTenantServiceClient(tenantConn)
	matchingClient := rentrelaypb.NewMatchingServiceClient(matchingConn)

	_, err := propertyClient.RegisterProperty(ctx, &rentrelaypb.RegisterPropertyRequest{
		LandlordId:  "landlord-matching-smoke",
		Title:       fmt.Sprintf("Matching Smoke Property %d", time.Now().UnixNano()),
		Address:     "Koramangala, Bengaluru",
		City:        "Bengaluru",
		Zone:        "south",
		Bedrooms:    2,
		RentMonthly: 25000,
		DepositAmt:  75000,
		Furnishing:  rentrelaypb.FurnishingType_SEMI_FURNISHED,
	})
	if err != nil {
		log.Fatalf("register property: %v", err)
	}

	tenantID := fmt.Sprintf("tenant-matching-smoke-%d", time.Now().UnixNano())
	rental, err := tenantClient.CreateRentalRequest(ctx, &rentrelaypb.CreateRentalRequestReq{
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

	resp, err := matchingClient.FindMatches(ctx, &rentrelaypb.MatchRequest{
		MatchRequestId: "smoke-" + rental.RequestId,
		RentalRequest:  rental,
	})
	if err != nil {
		log.Fatalf("find matches: %v", err)
	}

	fmt.Printf("match_request_id=%s candidates=%d\n", resp.MatchRequestId, len(resp.Candidates))
	if len(resp.Candidates) > 0 {
		best := resp.Candidates[0]
		fmt.Printf("best property_id=%s score=%.3f reason=%s\n", best.PropertyId, best.Score, best.MatchReason)
	}
}

func mustConnect(addr string) *grpc.ClientConn {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("create grpc client for %s: %v", addr, err)
	}
	return conn
}

func envOrDefault(name string, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
