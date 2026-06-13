package agreement

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
	rentrelaypb.UnimplementedAgreementServiceServer

	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func NewInMemoryService() *Service {
	return NewService(NewMemoryRepository())
}

func (s *Service) CreateAgreement(ctx context.Context, req *rentrelaypb.CreateAgreementRequest) (*rentrelaypb.Agreement, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	if strings.TrimSpace(req.TenantId) == "" || strings.TrimSpace(req.LandlordId) == "" || strings.TrimSpace(req.PropertyId) == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id, landlord_id, and property_id are required")
	}
	if req.MonthlyRent <= 0 || req.DepositAmount < 0 || req.LeaseMonths <= 0 || req.NoticeDays <= 0 {
		return nil, status.Error(codes.InvalidArgument, "rent and lease terms are invalid")
	}

	now := timestamppb.Now()
	startDate := req.StartDate
	if startDate == nil {
		startDate = now
	}
	endDate := timestamppb.New(startDate.AsTime().AddDate(0, int(req.LeaseMonths), 0))

	agreement := &rentrelaypb.Agreement{
		AgreementId:    newID("agreement"),
		TenantId:       strings.TrimSpace(req.TenantId),
		LandlordId:     strings.TrimSpace(req.LandlordId),
		PropertyId:     strings.TrimSpace(req.PropertyId),
		State:          rentrelaypb.AgreementState_CREATED,
		MonthlyRent:    req.MonthlyRent,
		DepositAmount:  req.DepositAmount,
		DepositHeld:    0,
		LeaseMonths:    req.LeaseMonths,
		NoticeDays:     req.NoticeDays,
		StartDate:      startDate,
		EndDate:        endDate,
		CreatedAt:      now,
		UpdatedAt:      now,
		ReplicaVersion: 1,
	}

	if err := s.repo.Create(ctx, agreement); err != nil {
		return nil, status.Errorf(codes.Internal, "create agreement: %v", err)
	}
	return cloneAgreement(agreement), nil
}

func (s *Service) GetAgreement(ctx context.Context, req *rentrelaypb.AgreementActionRequest) (*rentrelaypb.Agreement, error) {
	if req == nil || strings.TrimSpace(req.AgreementId) == "" {
		return nil, status.Error(codes.InvalidArgument, "agreement_id is required")
	}
	return s.findAgreement(ctx, req.AgreementId)
}

func (s *Service) SignAgreement(ctx context.Context, req *rentrelaypb.SignAgreementRequest) (*rentrelaypb.Agreement, error) {
	if req == nil || strings.TrimSpace(req.AgreementId) == "" || strings.TrimSpace(req.SignerId) == "" || strings.TrimSpace(req.SignatureHash) == "" {
		return nil, status.Error(codes.InvalidArgument, "agreement_id, signer_id, and signature_hash are required")
	}

	agreement, err := s.findAgreement(ctx, req.AgreementId)
	if err != nil {
		return nil, err
	}
	if agreement.State != rentrelaypb.AgreementState_CREATED {
		return nil, invalidTransition(agreement.State, "sign")
	}
	if req.SignerId != agreement.TenantId && req.SignerId != agreement.LandlordId {
		return nil, status.Error(codes.PermissionDenied, "signer is not a party to the agreement")
	}

	signature := req.SignerId + ":" + req.SignatureHash
	if !hasSigner(agreement.Signatures, req.SignerId) {
		agreement.Signatures = append(agreement.Signatures, signature)
	}
	if hasSigner(agreement.Signatures, agreement.TenantId) && hasSigner(agreement.Signatures, agreement.LandlordId) {
		agreement.State = rentrelaypb.AgreementState_SIGNED
	}

	return s.saveAgreement(ctx, agreement)
}

func (s *Service) HoldEscrow(ctx context.Context, req *rentrelaypb.AgreementActionRequest) (*rentrelaypb.Agreement, error) {
	agreement, err := s.actionAgreement(ctx, req)
	if err != nil {
		return nil, err
	}
	if agreement.State != rentrelaypb.AgreementState_SIGNED {
		return nil, invalidTransition(agreement.State, "hold escrow")
	}
	if req.ActorId != agreement.TenantId {
		return nil, status.Error(codes.PermissionDenied, "only the tenant can fund escrow")
	}

	agreement.DepositHeld = agreement.DepositAmount
	agreement.State = rentrelaypb.AgreementState_ESCROW_HELD
	return s.saveAgreement(ctx, agreement)
}

func (s *Service) StartLease(ctx context.Context, req *rentrelaypb.AgreementActionRequest) (*rentrelaypb.Agreement, error) {
	agreement, err := s.actionAgreement(ctx, req)
	if err != nil {
		return nil, err
	}
	if agreement.State != rentrelaypb.AgreementState_ESCROW_HELD {
		return nil, invalidTransition(agreement.State, "start lease")
	}
	if req.ActorId != agreement.LandlordId {
		return nil, status.Error(codes.PermissionDenied, "only the landlord can start the lease")
	}

	agreement.State = rentrelaypb.AgreementState_ACTIVE
	return s.saveAgreement(ctx, agreement)
}

func (s *Service) InitiateNotice(ctx context.Context, req *rentrelaypb.AgreementActionRequest) (*rentrelaypb.Agreement, error) {
	agreement, err := s.actionAgreement(ctx, req)
	if err != nil {
		return nil, err
	}
	if agreement.State != rentrelaypb.AgreementState_ACTIVE {
		return nil, invalidTransition(agreement.State, "initiate notice")
	}
	agreement.State = rentrelaypb.AgreementState_NOTICE_PERIOD
	return s.saveAgreement(ctx, agreement)
}

func (s *Service) VacateProperty(ctx context.Context, req *rentrelaypb.AgreementActionRequest) (*rentrelaypb.Agreement, error) {
	agreement, err := s.actionAgreement(ctx, req)
	if err != nil {
		return nil, err
	}
	if agreement.State != rentrelaypb.AgreementState_NOTICE_PERIOD {
		return nil, invalidTransition(agreement.State, "vacate property")
	}
	agreement.State = rentrelaypb.AgreementState_TERMINATING
	return s.saveAgreement(ctx, agreement)
}

func (s *Service) ReleaseEscrow(ctx context.Context, req *rentrelaypb.ReleaseEscrowRequest) (*rentrelaypb.Agreement, error) {
	if req == nil || strings.TrimSpace(req.AgreementId) == "" || strings.TrimSpace(req.AuthorizedBy) == "" {
		return nil, status.Error(codes.InvalidArgument, "agreement_id and authorized_by are required")
	}
	agreement, err := s.findAgreement(ctx, req.AgreementId)
	if err != nil {
		return nil, err
	}
	if agreement.State != rentrelaypb.AgreementState_TERMINATING {
		return nil, invalidTransition(agreement.State, "release escrow")
	}
	if req.AuthorizedBy != agreement.LandlordId && req.AuthorizedBy != agreement.TenantId {
		return nil, status.Error(codes.PermissionDenied, "authorized user is not a party to the agreement")
	}
	if req.DeductionAmount < 0 || req.DeductionAmount > agreement.DepositHeld {
		return nil, status.Error(codes.InvalidArgument, "deduction_amount is invalid")
	}

	agreement.DepositHeld = 0
	agreement.State = rentrelaypb.AgreementState_COMPLETED
	return s.saveAgreement(ctx, agreement)
}

func (s *Service) RecordPayment(context.Context, *rentrelaypb.RecordPaymentReq) (*rentrelaypb.PaymentReceipt, error) {
	return nil, status.Error(codes.Unimplemented, "RecordPayment will be implemented with payment persistence")
}

func (s *Service) TransitionState(context.Context, *rentrelaypb.AgreementActionRequest) (*rentrelaypb.Agreement, error) {
	return nil, status.Error(codes.Unimplemented, "use the explicit lifecycle RPC methods")
}

func (s *Service) actionAgreement(ctx context.Context, req *rentrelaypb.AgreementActionRequest) (*rentrelaypb.Agreement, error) {
	if req == nil || strings.TrimSpace(req.AgreementId) == "" || strings.TrimSpace(req.ActorId) == "" {
		return nil, status.Error(codes.InvalidArgument, "agreement_id and actor_id are required")
	}
	return s.findAgreement(ctx, req.AgreementId)
}

func (s *Service) findAgreement(ctx context.Context, agreementID string) (*rentrelaypb.Agreement, error) {
	agreement, err := s.repo.FindByID(ctx, strings.TrimSpace(agreementID))
	if errors.Is(err, ErrAgreementNotFound) {
		return nil, status.Error(codes.NotFound, "agreement not found")
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "find agreement: %v", err)
	}
	return agreement, nil
}

func (s *Service) saveAgreement(ctx context.Context, agreement *rentrelaypb.Agreement) (*rentrelaypb.Agreement, error) {
	agreement.UpdatedAt = timestamppb.Now()
	agreement.ReplicaVersion++
	if err := s.repo.Save(ctx, agreement); err != nil {
		return nil, status.Errorf(codes.Internal, "save agreement: %v", err)
	}
	return cloneAgreement(agreement), nil
}

func hasSigner(signatures []string, signerID string) bool {
	prefix := signerID + ":"
	for _, signature := range signatures {
		if strings.HasPrefix(signature, prefix) {
			return true
		}
	}
	return false
}

func invalidTransition(state rentrelaypb.AgreementState, action string) error {
	return status.Errorf(codes.FailedPrecondition, "cannot %s while agreement is in state %s", action, state.String())
}

func newID(prefix string) string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
	}
	return fmt.Sprintf("%s-%s", prefix, hex.EncodeToString(b[:]))
}
