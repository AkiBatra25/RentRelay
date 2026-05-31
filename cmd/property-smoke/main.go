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
	addr := os.Getenv("PROPERTY_SERVICE_ADDR")
	if addr == "" {
		addr = "localhost:50052"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("create grpc client: %v", err)
	}
	defer conn.Close()

	client := rentrelaypb.NewPropertyServiceClient(conn)

	created, err := client.RegisterProperty(ctx, &rentrelaypb.RegisterPropertyRequest{
		LandlordId:  "landlord-smoke",
		Title:       "2BHK near HSR Layout",
		Address:     "27th Main Road, HSR Layout",
		City:        "Bengaluru",
		Zone:        "south",
		Latitude:    12.9116,
		Longitude:   77.6474,
		Bedrooms:    2,
		RentMonthly: 28000,
		DepositAmt:  84000,
		Furnishing:  rentrelaypb.FurnishingType_SEMI_FURNISHED,
		Amenities:   []string{"parking", "lift", "power backup"},
	})
	if err != nil {
		log.Fatalf("register property: %v", err)
	}

	searchResp, err := client.SearchProperties(ctx, &rentrelaypb.SearchPropertiesRequest{
		City:        "Bengaluru",
		Zone:        "south",
		MinBedrooms: 2,
		MaxRent:     30000,
		Furnishing:  rentrelaypb.FurnishingType_SEMI_FURNISHED,
	})
	if err != nil {
		log.Fatalf("search properties: %v", err)
	}

	updated, err := client.UpdateAvailability(ctx, &rentrelaypb.UpdateAvailabilityRequest{
		PropertyId:  created.PropertyId,
		IsAvailable: false,
	})
	if err != nil {
		log.Fatalf("update availability: %v", err)
	}

	fmt.Printf("registered property_id=%s title=%q\n", created.PropertyId, created.Title)
	fmt.Printf("search results=%d\n", len(searchResp.Properties))
	fmt.Printf("updated availability=%v\n", updated.IsAvailable)
}
