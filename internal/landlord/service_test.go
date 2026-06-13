package landlord

import (
	"context"
	"errors"
	"testing"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
	"google.golang.org/grpc"
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

func TestServiceRejectsLeaseTermsForAnotherLandlordsProperty(t *testing.T) {
	svc := NewService(NewMemoryRepository(), &fakePropertyClient{
		property: &rentrelaypb.Property{
			PropertyId:  "property-1",
			LandlordId:  "landlord-2",
			IsAvailable: true,
		},
	})

	_, err := svc.SetLeaseTerms(context.Background(), &rentrelaypb.SetLeaseTermsRequest{
		LandlordId: "landlord-1",
		PropertyId: "property-1",
		Terms: &rentrelaypb.LeaseTerms{
			LeaseDurationMo:  11,
			NoticePeriodDays: 30,
			PaymentDueDay:    "5",
		},
	})
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("SetLeaseTerms() code = %v, want %v", status.Code(err), codes.PermissionDenied)
	}
}

func TestServiceGetDashboard(t *testing.T) {
	svc := NewService(NewMemoryRepository(), &fakePropertyClient{
		properties: []*rentrelaypb.Property{
			{PropertyId: "property-1", LandlordId: "landlord-1", RentMonthly: 25000, IsAvailable: true},
			{PropertyId: "property-2", LandlordId: "landlord-1", RentMonthly: 30000, IsAvailable: false},
		},
	})

	dashboard, err := svc.GetDashboard(context.Background(), &rentrelaypb.LandlordDashboardRequest{
		LandlordId: "landlord-1",
	})
	if err != nil {
		t.Fatalf("GetDashboard() error = %v", err)
	}

	if dashboard.TotalProperties != 2 {
		t.Fatalf("TotalProperties = %d, want 2", dashboard.TotalProperties)
	}
	if dashboard.TotalRentThisMonth != 25000 {
		t.Fatalf("TotalRentThisMonth = %.2f, want 25000", dashboard.TotalRentThisMonth)
	}
}

type fakePropertyClient struct {
	property   *rentrelaypb.Property
	properties []*rentrelaypb.Property
	err        error
}

func (c *fakePropertyClient) RegisterProperty(ctx context.Context, in *rentrelaypb.RegisterPropertyRequest, opts ...grpc.CallOption) (*rentrelaypb.Property, error) {
	return nil, errors.New("not implemented")
}

func (c *fakePropertyClient) GetProperty(ctx context.Context, in *rentrelaypb.GetPropertyRequest, opts ...grpc.CallOption) (*rentrelaypb.Property, error) {
	if c.err != nil {
		return nil, c.err
	}
	return c.property, nil
}

func (c *fakePropertyClient) SearchProperties(ctx context.Context, in *rentrelaypb.SearchPropertiesRequest, opts ...grpc.CallOption) (*rentrelaypb.SearchPropertiesResponse, error) {
	return nil, errors.New("not implemented")
}

func (c *fakePropertyClient) UpdateAvailability(ctx context.Context, in *rentrelaypb.UpdateAvailabilityRequest, opts ...grpc.CallOption) (*rentrelaypb.Property, error) {
	return nil, errors.New("not implemented")
}

func (c *fakePropertyClient) ListByLandlord(ctx context.Context, in *rentrelaypb.GetUserRequest, opts ...grpc.CallOption) (*rentrelaypb.SearchPropertiesResponse, error) {
	if c.err != nil {
		return nil, c.err
	}
	return &rentrelaypb.SearchPropertiesResponse{Properties: c.properties}, nil
}
