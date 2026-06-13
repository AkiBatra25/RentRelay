package tenant

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	rentrelaypb.UnimplementedTenantServiceServer

	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func NewInMemoryService() *Service {
	return NewService(NewMemoryRepository())
}

func (s *Service) CreateRentalRequest(ctx context.Context, req *rentrelaypb.CreateRentalRequestReq) (*rentrelaypb.RentalRequest, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	if strings.TrimSpace(req.TenantId) == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}
	if strings.TrimSpace(req.PreferredCity) == "" {
		return nil, status.Error(codes.InvalidArgument, "preferred_city is required")
	}
	if req.BedroomsNeeded <= 0 {
		return nil, status.Error(codes.InvalidArgument, "bedrooms_needed must be greater than zero")
	}
	if req.MaxRent <= 0 {
		return nil, status.Error(codes.InvalidArgument, "max_rent must be greater than zero")
	}

	request := &rentrelaypb.RentalRequest{
		RequestId:      newID("rental-request"),
		TenantId:       strings.TrimSpace(req.TenantId),
		PreferredZone:  strings.TrimSpace(req.PreferredZone),
		PreferredCity:  strings.TrimSpace(req.PreferredCity),
		BedroomsNeeded: req.BedroomsNeeded,
		MaxRent:        req.MaxRent,
		Furnishing:     req.Furnishing,
		MoveInDate:     req.MoveInDate,
		CreatedAt:      timestamppb.Now(),
	}
	if request.MoveInDate == nil {
		request.MoveInDate = timestamppb.Now()
	}

	if err := s.repo.CreateRentalRequest(ctx, request); err != nil {
		return nil, status.Errorf(codes.Internal, "create rental request: %v", err)
	}

	return cloneRentalRequest(request), nil
}

func (s *Service) GetRentalRequest(ctx context.Context, req *rentrelaypb.GetUserRequest) (*rentrelaypb.RentalRequest, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	request, err := s.repo.FindRentalRequestByTenant(ctx, strings.TrimSpace(req.UserId))
	if err != nil {
		if errors.Is(err, ErrRentalRequestNotFound) {
			return nil, status.Error(codes.NotFound, "rental request not found")
		}
		return nil, status.Errorf(codes.Internal, "find rental request: %v", err)
	}

	return request, nil
}

func (s *Service) GetDashboard(ctx context.Context, req *rentrelaypb.GetUserRequest) (*rentrelaypb.TenantDashboard, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	tenantID := strings.TrimSpace(req.UserId)
	if tenantID == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}

	return &rentrelaypb.TenantDashboard{
		TenantId:          tenantID,
		ActiveAgreementId: "",
		NextRentDue:       0,
		NextDueDate:       nil,
		MonthsPaid:        0,
		PendingDisputes:   0,
	}, nil
}

func (s *Service) InitiatePayment(ctx context.Context, req *rentrelaypb.PaymentRequest) (*rentrelaypb.PaymentReceipt, error) {
	return nil, status.Error(codes.Unimplemented, "InitiatePayment will be implemented after AgreementService")
}

func (s *Service) RaiseDispute(ctx context.Context, req *rentrelaypb.DisputeRequest) (*rentrelaypb.Dispute, error) {
	return nil, status.Error(codes.Unimplemented, "RaiseDispute will be implemented after AgreementService")
}

func newID(prefix string) string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
	}
	return fmt.Sprintf("%s-%s", prefix, hex.EncodeToString(b[:]))
}
