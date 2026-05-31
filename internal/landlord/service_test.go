package landlord

import (
	"context"
	"testing"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServiceSetAndGetLeaseTerms(t *testing.T) {
	svc := NewInMemoryService(nil)

	created, err := svc.SetLeaseTerms(context.Background(), &rentrelaypb.SetLeaseTermsRequest{
		LandlordId: "landlord-1",
		PropertyId: "property-1",
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
		t.Fatalf("SetLeaseTerms() error = %v", err)
	}

	if created.LandlordId != "landlord-1" {
		t.Fatalf("LandlordId = %q, want landlord-1", created.LandlordId)
	}
	if created.PropertyId != "property-1" {
		t.Fatalf("PropertyId = %q, want property-1", created.PropertyId)
	}

	got, err := svc.GetLeaseTerms(context.Background(), &rentrelaypb.GetLeaseTermsRequest{
		LandlordId: "landlord-1",
		PropertyId: "property-1",
	})
	if err != nil {
		t.Fatalf("GetLeaseTerms() error = %v", err)
	}

	if got.LeaseDurationMo != 11 {
		t.Fatalf("LeaseDurationMo = %d, want 11", got.LeaseDurationMo)
	}
}

func TestServiceGetLeaseTermsNotFound(t *testing.T) {
	svc := NewInMemoryService(nil)

	_, err := svc.GetLeaseTerms(context.Background(), &rentrelaypb.GetLeaseTermsRequest{
		LandlordId: "landlord-1",
		PropertyId: "missing",
	})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("GetLeaseTerms() code = %v, want %v", status.Code(err), codes.NotFound)
	}
}

func TestServiceRejectsInvalidLeaseTerms(t *testing.T) {
	svc := NewInMemoryService(nil)

	_, err := svc.SetLeaseTerms(context.Background(), &rentrelaypb.SetLeaseTermsRequest{
		LandlordId: "landlord-1",
		PropertyId: "property-1",
		Terms: &rentrelaypb.LeaseTerms{
			LeaseDurationMo:  0,
			NoticePeriodDays: 30,
			PaymentDueDay:    "5",
		},
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("SetLeaseTerms() code = %v, want %v", status.Code(err), codes.InvalidArgument)
	}
}
