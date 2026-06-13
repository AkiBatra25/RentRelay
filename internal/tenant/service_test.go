package tenant

import (
	"context"
	"testing"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServiceCreateAndGetRentalRequest(t *testing.T) {
	svc := NewInMemoryService()

	created, err := svc.CreateRentalRequest(context.Background(), &rentrelaypb.CreateRentalRequestReq{
		TenantId:       "tenant-1",
		PreferredCity:  "Bengaluru",
		PreferredZone:  "south",
		BedroomsNeeded: 2,
		MaxRent:        30000,
		Furnishing:     rentrelaypb.FurnishingType_SEMI_FURNISHED,
	})
	if err != nil {
		t.Fatalf("CreateRentalRequest() error = %v", err)
	}
	if created.RequestId == "" {
		t.Fatal("CreateRentalRequest() request_id is empty")
	}

	got, err := svc.GetRentalRequest(context.Background(), &rentrelaypb.GetUserRequest{UserId: "tenant-1"})
	if err != nil {
		t.Fatalf("GetRentalRequest() error = %v", err)
	}
	if got.RequestId != created.RequestId {
		t.Fatalf("GetRentalRequest() request_id = %q, want %q", got.RequestId, created.RequestId)
	}
}

func TestServiceGetRentalRequestNotFound(t *testing.T) {
	svc := NewInMemoryService()

	_, err := svc.GetRentalRequest(context.Background(), &rentrelaypb.GetUserRequest{UserId: "missing"})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("GetRentalRequest() code = %v, want %v", status.Code(err), codes.NotFound)
	}
}

func TestServiceRejectsInvalidRentalRequest(t *testing.T) {
	svc := NewInMemoryService()

	_, err := svc.CreateRentalRequest(context.Background(), &rentrelaypb.CreateRentalRequestReq{
		TenantId:       "tenant-1",
		PreferredCity:  "Bengaluru",
		BedroomsNeeded: 0,
		MaxRent:        30000,
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("CreateRentalRequest() code = %v, want %v", status.Code(err), codes.InvalidArgument)
	}
}

func TestServiceGetDashboard(t *testing.T) {
	svc := NewInMemoryService()

	dashboard, err := svc.GetDashboard(context.Background(), &rentrelaypb.GetUserRequest{UserId: "tenant-1"})
	if err != nil {
		t.Fatalf("GetDashboard() error = %v", err)
	}
	if dashboard.TenantId != "tenant-1" {
		t.Fatalf("TenantId = %q, want tenant-1", dashboard.TenantId)
	}
}
