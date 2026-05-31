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
	landlordAddr := os.Getenv("LANDLORD_SERVICE_ADDR")
	if landlordAddr == "" {
		landlordAddr = "localhost:50053"
	}

	propertyAddr := os.Getenv("PROPERTY_SERVICE_ADDR")
	if propertyAddr == "" {
		propertyAddr = "localhost:50052"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	landlordConn, err := grpc.NewClient(landlordAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("create landlord grpc client: %v", err)
	}
	defer landlordConn.Close()

	propertyConn, err := grpc.NewClient(propertyAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("create property grpc client: %v", err)
	}
	defer propertyConn.Close()

	landlordClient := rentrelaypb.NewLandlordServiceClient(landlordConn)
	propertyClient := rentrelaypb.NewPropertyServiceClient(propertyConn)

	createdProperty, err := propertyClient.RegisterProperty(ctx, &rentrelaypb.RegisterPropertyRequest{
		LandlordId:  "landlord-smoke",
		Title:       fmt.Sprintf("Smoke Property %d", time.Now().UnixNano()),
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

	terms, err := landlordClient.SetLeaseTerms(ctx, &rentrelaypb.SetLeaseTermsRequest{
		LandlordId: "landlord-smoke",
		PropertyId: createdProperty.PropertyId,
		Terms: &rentrelaypb.LeaseTerms{
			LeaseDurationMo:   11,
			NoticePeriodDays:  30,
			PreferredTenant:   "family",
			AllowedTypes:      []string{"family", "working professional"},
			MaintenanceCharge: 2500,
			PaymentDueDay:     "5",
		},
	})
	if err != nil {
		log.Fatalf("set lease terms: %v", err)
	}

	fetched, err := landlordClient.GetLeaseTerms(ctx, &rentrelaypb.GetLeaseTermsRequest{
		LandlordId: "landlord-smoke",
		PropertyId: createdProperty.PropertyId,
	})
	if err != nil {
		log.Fatalf("get lease terms: %v", err)
	}

	dashboard, err := landlordClient.GetDashboard(ctx, &rentrelaypb.LandlordDashboardRequest{
		LandlordId: "landlord-smoke",
	})
	if err != nil {
		log.Fatalf("get dashboard: %v", err)
	}

	fmt.Printf("registered property_id=%s title=%q\n", createdProperty.PropertyId, createdProperty.Title)
	fmt.Printf("set lease_terms property_id=%s duration=%d months\n", terms.PropertyId, terms.LeaseDurationMo)
	fmt.Printf("fetched lease_terms payment_due_day=%s\n", fetched.PaymentDueDay)
	fmt.Printf("dashboard total_properties=%d total_rent_this_month=%.2f\n", dashboard.TotalProperties, dashboard.TotalRentThisMonth)
}
