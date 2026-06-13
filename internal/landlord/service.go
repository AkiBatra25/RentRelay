package landlord

import (
	"context"
	"errors"
	"strings"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	rentrelaypb.UnimplementedLandlordServiceServer

	repo           Repository
	propertyClient rentrelaypb.PropertyServiceClient
}

func NewService(repo Repository, propertyClient rentrelaypb.PropertyServiceClient) *Service {
	return &Service{
		repo:           repo,
		propertyClient: propertyClient,
	}
}

func NewInMemoryService(propertyClient rentrelaypb.PropertyServiceClient) *Service {
	return NewService(NewMemoryRepository(), propertyClient)
}

func (s *Service) SetLeaseTerms(ctx context.Context, req *rentrelaypb.SetLeaseTermsRequest) (*rentrelaypb.LeaseTerms, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	if strings.TrimSpace(req.LandlordId) == "" {
		return nil, status.Error(codes.InvalidArgument, "landlord_id is required")
	}
	if strings.TrimSpace(req.PropertyId) == "" {
		return nil, status.Error(codes.InvalidArgument, "property_id is required")
	}
	if req.Terms == nil {
		return nil, status.Error(codes.InvalidArgument, "terms are required")
	}

	terms := cloneLeaseTerms(req.Terms)
	terms.LandlordId = strings.TrimSpace(req.LandlordId)
	terms.PropertyId = strings.TrimSpace(req.PropertyId)

	if terms.LeaseDurationMo <= 0 {
		return nil, status.Error(codes.InvalidArgument, "lease_duration_mo must be greater than zero")
	}
	if terms.NoticePeriodDays <= 0 {
		return nil, status.Error(codes.InvalidArgument, "notice_period_days must be greater than zero")
	}
	if strings.TrimSpace(terms.PaymentDueDay) == "" {
		return nil, status.Error(codes.InvalidArgument, "payment_due_day is required")
	}
	if err := s.ensureLandlordOwnsProperty(ctx, terms.LandlordId, terms.PropertyId); err != nil {
		return nil, err
	}

	if err := s.repo.SaveLeaseTerms(ctx, terms); err != nil {
		return nil, status.Errorf(codes.Internal, "save lease terms: %v", err)
	}

	return cloneLeaseTerms(terms), nil
}

func (s *Service) GetLeaseTerms(ctx context.Context, req *rentrelaypb.GetLeaseTermsRequest) (*rentrelaypb.LeaseTerms, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	landlordID := strings.TrimSpace(req.LandlordId)
	propertyID := strings.TrimSpace(req.PropertyId)
	if landlordID == "" {
		return nil, status.Error(codes.InvalidArgument, "landlord_id is required")
	}
	if propertyID == "" {
		return nil, status.Error(codes.InvalidArgument, "property_id is required")
	}

	terms, err := s.repo.FindLeaseTerms(ctx, landlordID, propertyID)
	if err != nil {
		if errors.Is(err, ErrLeaseTermsNotFound) {
			return nil, status.Error(codes.NotFound, "lease terms not found")
		}
		return nil, status.Errorf(codes.Internal, "find lease terms: %v", err)
	}

	return terms, nil
}

func (s *Service) GetDashboard(ctx context.Context, req *rentrelaypb.LandlordDashboardRequest) (*rentrelaypb.LandlordDashboard, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	landlordID := strings.TrimSpace(req.LandlordId)
	if landlordID == "" {
		return nil, status.Error(codes.InvalidArgument, "landlord_id is required")
	}
	if s.propertyClient == nil {
		return nil, status.Error(codes.FailedPrecondition, "property service client is not configured")
	}

	propertiesResp, err := s.propertyClient.ListByLandlord(ctx, &rentrelaypb.GetUserRequest{
		UserId: landlordID,
	})
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "list landlord properties: %v", err)
	}

	var totalRent float64
	for _, property := range propertiesResp.Properties {
		if property.IsAvailable {
			totalRent += property.RentMonthly
		}
	}

	return &rentrelaypb.LandlordDashboard{
		LandlordId:         landlordID,
		TotalProperties:    int32(len(propertiesResp.Properties)),
		ActiveLeases:       0,
		PendingDisputes:    0,
		TotalRentThisMonth: totalRent,
		Properties:         propertiesResp.Properties,
	}, nil
}

func (s *Service) RaiseDispute(ctx context.Context, req *rentrelaypb.DisputeRequest) (*rentrelaypb.Dispute, error) {
	return nil, status.Error(codes.Unimplemented, "RaiseDispute will be implemented after AgreementService")
}

func (s *Service) ConfirmVacation(ctx context.Context, req *rentrelaypb.AgreementActionRequest) (*rentrelaypb.Agreement, error) {
	return nil, status.Error(codes.Unimplemented, "ConfirmVacation will be implemented after AgreementService")
}

func (s *Service) ensureLandlordOwnsProperty(ctx context.Context, landlordID string, propertyID string) error {
	if s.propertyClient == nil {
		return nil
	}

	property, err := s.propertyClient.GetProperty(ctx, &rentrelaypb.GetPropertyRequest{PropertyId: propertyID})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return status.Error(codes.NotFound, "property not found")
		}
		return status.Errorf(codes.Unavailable, "get property: %v", err)
	}
	if property == nil {
		return status.Error(codes.Unavailable, "property service returned an empty property")
	}
	if property.LandlordId != landlordID {
		return status.Error(codes.PermissionDenied, "property does not belong to landlord")
	}

	return nil
}
