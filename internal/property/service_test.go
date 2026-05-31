package property

import (
	"context"
	"testing"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
)

func TestServiceRegisterAndGetProperty(t *testing.T) {
	svc := NewInMemoryService()

	created, err := svc.RegisterProperty(context.Background(), &rentrelaypb.RegisterPropertyRequest{
		LandlordId:  "landlord-1",
		Title:       "2BHK near metro",
		Address:     "Test address",
		City:        "Bengaluru",
		Zone:        "south",
		Latitude:    12.9716,
		Longitude:   77.5946,
		Bedrooms:    2,
		RentMonthly: 25000,
		DepositAmt:  75000,
		Furnishing:  rentrelaypb.FurnishingType_SEMI_FURNISHED,
		Amenities:   []string{"parking", "lift"},
	})
	if err != nil {
		t.Fatalf("RegisterProperty() error = %v", err)
	}

	if created.PropertyId == "" {
		t.Fatal("RegisterProperty() property_id is empty")
	}

	got, err := svc.GetProperty(context.Background(), &rentrelaypb.GetPropertyRequest{
		PropertyId: created.PropertyId,
	})
	if err != nil {
		t.Fatalf("GetProperty() error = %v", err)
	}

	if got.PropertyId != created.PropertyId {
		t.Fatalf("GetProperty() property_id = %q, want %q", got.PropertyId, created.PropertyId)
	}
}
