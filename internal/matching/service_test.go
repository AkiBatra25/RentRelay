package matching

import (
	"context"
	"errors"
	"testing"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
	"google.golang.org/grpc"
)

func TestFindMatchesScoresAndSortsCandidates(t *testing.T) {
	svc := NewService(
		&fakePropertyClient{properties: []*rentrelaypb.Property{
			{
				PropertyId:  "property-expensive",
				LandlordId:  "landlord-2",
				City:        "Bengaluru",
				Zone:        "south",
				Bedrooms:    2,
				RentMonthly: 29000,
				Furnishing:  rentrelaypb.FurnishingType_SEMI_FURNISHED,
				IsAvailable: true,
			},
			{
				PropertyId:  "property-best",
				LandlordId:  "landlord-1",
				City:        "Bengaluru",
				Zone:        "south",
				Bedrooms:    2,
				RentMonthly: 24000,
				Furnishing:  rentrelaypb.FurnishingType_SEMI_FURNISHED,
				IsAvailable: true,
			},
		}},
		&fakeLandlordClient{terms: &rentrelaypb.LeaseTerms{LeaseDurationMo: 11}},
		nil,
	)

	resp, err := svc.FindMatches(context.Background(), &rentrelaypb.MatchRequest{
		MatchRequestId: "match-1",
		RentalRequest: &rentrelaypb.RentalRequest{
			RequestId:      "request-1",
			TenantId:       "tenant-1",
			PreferredCity:  "Bengaluru",
			PreferredZone:  "south",
			BedroomsNeeded: 2,
			MaxRent:        30000,
			Furnishing:     rentrelaypb.FurnishingType_SEMI_FURNISHED,
		},
	})
	if err != nil {
		t.Fatalf("FindMatches() error = %v", err)
	}
	if len(resp.Candidates) != 2 {
		t.Fatalf("FindMatches() candidates = %d, want 2", len(resp.Candidates))
	}
	if resp.Candidates[0].PropertyId != "property-best" {
		t.Fatalf("first candidate = %q, want property-best", resp.Candidates[0].PropertyId)
	}
	if resp.Candidates[0].Terms == nil {
		t.Fatal("first candidate lease terms are nil")
	}
}

func TestFindMatchesRequiresRentalRequest(t *testing.T) {
	svc := NewService(&fakePropertyClient{}, nil, nil)

	if _, err := svc.FindMatches(context.Background(), &rentrelaypb.MatchRequest{}); err == nil {
		t.Fatal("FindMatches() error = nil, want validation error")
	}
}

type fakePropertyClient struct {
	properties []*rentrelaypb.Property
	property   *rentrelaypb.Property
	err        error
}

func (c *fakePropertyClient) RegisterProperty(context.Context, *rentrelaypb.RegisterPropertyRequest, ...grpc.CallOption) (*rentrelaypb.Property, error) {
	return nil, errors.New("not implemented")
}

func (c *fakePropertyClient) GetProperty(context.Context, *rentrelaypb.GetPropertyRequest, ...grpc.CallOption) (*rentrelaypb.Property, error) {
	if c.err != nil {
		return nil, c.err
	}
	return c.property, nil
}

func (c *fakePropertyClient) SearchProperties(context.Context, *rentrelaypb.SearchPropertiesRequest, ...grpc.CallOption) (*rentrelaypb.SearchPropertiesResponse, error) {
	if c.err != nil {
		return nil, c.err
	}
	return &rentrelaypb.SearchPropertiesResponse{Properties: c.properties}, nil
}

func (c *fakePropertyClient) UpdateAvailability(context.Context, *rentrelaypb.UpdateAvailabilityRequest, ...grpc.CallOption) (*rentrelaypb.Property, error) {
	if c.err != nil {
		return nil, c.err
	}
	return c.property, nil
}

func (c *fakePropertyClient) ListByLandlord(context.Context, *rentrelaypb.GetUserRequest, ...grpc.CallOption) (*rentrelaypb.SearchPropertiesResponse, error) {
	return nil, errors.New("not implemented")
}

type fakeLandlordClient struct {
	terms *rentrelaypb.LeaseTerms
	err   error
}

func (c *fakeLandlordClient) SetLeaseTerms(context.Context, *rentrelaypb.SetLeaseTermsRequest, ...grpc.CallOption) (*rentrelaypb.LeaseTerms, error) {
	return nil, errors.New("not implemented")
}

func (c *fakeLandlordClient) GetLeaseTerms(context.Context, *rentrelaypb.GetLeaseTermsRequest, ...grpc.CallOption) (*rentrelaypb.LeaseTerms, error) {
	if c.err != nil {
		return nil, c.err
	}
	return c.terms, nil
}

func (c *fakeLandlordClient) GetDashboard(context.Context, *rentrelaypb.LandlordDashboardRequest, ...grpc.CallOption) (*rentrelaypb.LandlordDashboard, error) {
	return nil, errors.New("not implemented")
}

func (c *fakeLandlordClient) RaiseDispute(context.Context, *rentrelaypb.DisputeRequest, ...grpc.CallOption) (*rentrelaypb.Dispute, error) {
	return nil, errors.New("not implemented")
}

func (c *fakeLandlordClient) ConfirmVacation(context.Context, *rentrelaypb.AgreementActionRequest, ...grpc.CallOption) (*rentrelaypb.Agreement, error) {
	return nil, errors.New("not implemented")
}

type fakeAgreementClient struct {
	agreement *rentrelaypb.Agreement
	err       error
}

func (c *fakeAgreementClient) CreateAgreement(context.Context, *rentrelaypb.CreateAgreementRequest, ...grpc.CallOption) (*rentrelaypb.Agreement, error) {
	if c.err != nil {
		return nil, c.err
	}
	return c.agreement, nil
}

func (c *fakeAgreementClient) GetAgreement(context.Context, *rentrelaypb.AgreementActionRequest, ...grpc.CallOption) (*rentrelaypb.Agreement, error) {
	return nil, errors.New("not implemented")
}

func (c *fakeAgreementClient) SignAgreement(context.Context, *rentrelaypb.SignAgreementRequest, ...grpc.CallOption) (*rentrelaypb.Agreement, error) {
	return nil, errors.New("not implemented")
}

func (c *fakeAgreementClient) HoldEscrow(context.Context, *rentrelaypb.AgreementActionRequest, ...grpc.CallOption) (*rentrelaypb.Agreement, error) {
	return nil, errors.New("not implemented")
}

func (c *fakeAgreementClient) StartLease(context.Context, *rentrelaypb.AgreementActionRequest, ...grpc.CallOption) (*rentrelaypb.Agreement, error) {
	return nil, errors.New("not implemented")
}

func (c *fakeAgreementClient) RecordPayment(context.Context, *rentrelaypb.RecordPaymentReq, ...grpc.CallOption) (*rentrelaypb.PaymentReceipt, error) {
	return nil, errors.New("not implemented")
}

func (c *fakeAgreementClient) InitiateNotice(context.Context, *rentrelaypb.AgreementActionRequest, ...grpc.CallOption) (*rentrelaypb.Agreement, error) {
	return nil, errors.New("not implemented")
}

func (c *fakeAgreementClient) VacateProperty(context.Context, *rentrelaypb.AgreementActionRequest, ...grpc.CallOption) (*rentrelaypb.Agreement, error) {
	return nil, errors.New("not implemented")
}

func (c *fakeAgreementClient) ReleaseEscrow(context.Context, *rentrelaypb.ReleaseEscrowRequest, ...grpc.CallOption) (*rentrelaypb.Agreement, error) {
	return nil, errors.New("not implemented")
}

func (c *fakeAgreementClient) TransitionState(context.Context, *rentrelaypb.AgreementActionRequest, ...grpc.CallOption) (*rentrelaypb.Agreement, error) {
	return nil, errors.New("not implemented")
}

func (c *fakeAgreementClient) StreamAgreementEvents(context.Context, *rentrelaypb.AgreementActionRequest, ...grpc.CallOption) (grpc.ServerStreamingClient[rentrelaypb.AgreementEvent], error) {
	return nil, errors.New("not implemented")
}
