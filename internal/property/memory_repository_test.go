package property

import (
	"context"
	"errors"
	"testing"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
)

func TestMemoryRepositoryCreateAndFindByID(t *testing.T) {
	repo := NewMemoryRepository()

	property := testProperty("prop-1", "landlord-1")
	if err := repo.Create(context.Background(), property); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repo.FindByID(context.Background(), "prop-1")
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}

	if got.PropertyId != "prop-1" {
		t.Fatalf("FindByID() property_id = %q, want prop-1", got.PropertyId)
	}
}

func TestMemoryRepositoryRejectsDuplicatePropertyID(t *testing.T) {
	repo := NewMemoryRepository()

	property := testProperty("prop-1", "landlord-1")
	if err := repo.Create(context.Background(), property); err != nil {
		t.Fatalf("Create() first call error = %v", err)
	}

	err := repo.Create(context.Background(), property)
	if !errors.Is(err, ErrDuplicateProperty) {
		t.Fatalf("Create() duplicate error = %v, want %v", err, ErrDuplicateProperty)
	}
}

func TestMemoryRepositorySearchFiltersAvailableProperties(t *testing.T) {
	repo := NewMemoryRepository()

	availableMatch := testProperty("prop-1", "landlord-1")
	availableMatch.City = "Bengaluru"
	availableMatch.Zone = "south"
	availableMatch.Bedrooms = 2
	availableMatch.RentMonthly = 25000
	availableMatch.Furnishing = rentrelaypb.FurnishingType_SEMI_FURNISHED
	availableMatch.IsAvailable = true

	expensive := testProperty("prop-2", "landlord-2")
	expensive.City = "Bengaluru"
	expensive.Zone = "south"
	expensive.Bedrooms = 2
	expensive.RentMonthly = 50000
	expensive.Furnishing = rentrelaypb.FurnishingType_SEMI_FURNISHED
	expensive.IsAvailable = true

	unavailable := testProperty("prop-3", "landlord-3")
	unavailable.City = "Bengaluru"
	unavailable.Zone = "south"
	unavailable.Bedrooms = 2
	unavailable.RentMonthly = 24000
	unavailable.Furnishing = rentrelaypb.FurnishingType_SEMI_FURNISHED
	unavailable.IsAvailable = false

	for _, property := range []*rentrelaypb.Property{availableMatch, expensive, unavailable} {
		if err := repo.Create(context.Background(), property); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	results, err := repo.Search(context.Background(), SearchFilter{
		City:        "bengaluru",
		Zone:        "SOUTH",
		MinBedrooms: 2,
		MaxRent:     30000,
		Furnishing:  rentrelaypb.FurnishingType_SEMI_FURNISHED,
	})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Search() returned %d results, want 1", len(results))
	}
	if results[0].PropertyId != "prop-1" {
		t.Fatalf("Search() property_id = %q, want prop-1", results[0].PropertyId)
	}
}

func TestMemoryRepositoryUpdateAvailability(t *testing.T) {
	repo := NewMemoryRepository()

	property := testProperty("prop-1", "landlord-1")
	property.IsAvailable = true

	if err := repo.Create(context.Background(), property); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	updated, err := repo.UpdateAvailability(context.Background(), "prop-1", false)
	if err != nil {
		t.Fatalf("UpdateAvailability() error = %v", err)
	}

	if updated.IsAvailable {
		t.Fatal("UpdateAvailability() IsAvailable = true, want false")
	}
}

func TestMemoryRepositoryListByLandlord(t *testing.T) {
	repo := NewMemoryRepository()

	properties := []*rentrelaypb.Property{
		testProperty("prop-1", "landlord-1"),
		testProperty("prop-2", "landlord-1"),
		testProperty("prop-3", "landlord-2"),
	}

	for _, property := range properties {
		if err := repo.Create(context.Background(), property); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	results, err := repo.ListByLandlord(context.Background(), "landlord-1")
	if err != nil {
		t.Fatalf("ListByLandlord() error = %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("ListByLandlord() returned %d results, want 2", len(results))
	}
}

func testProperty(propertyID string, landlordID string) *rentrelaypb.Property {
	return &rentrelaypb.Property{
		PropertyId:  propertyID,
		LandlordId:  landlordID,
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
		IsAvailable: true,
	}
}
