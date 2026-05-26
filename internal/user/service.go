package user

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
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
	rentrelaypb.UnimplementedUserServiceServer

	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func NewInMemoryService() *Service {
	return NewService(NewMemoryRepository())
}

func (s *Service) Register(ctx context.Context, req *rentrelaypb.RegisterRequest) (*rentrelaypb.RegisterResponse, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if strings.TrimSpace(req.Email) == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	if strings.TrimSpace(req.Password) == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}
	if req.Role == rentrelaypb.UserRole_ROLE_UNKNOWN {
		return nil, status.Error(codes.InvalidArgument, "role is required")
	}

	now := timestamppb.Now()
	user := &rentrelaypb.User{
		UserId:      newID("user"),
		Name:        strings.TrimSpace(req.Name),
		Email:       normalizeEmail(req.Email),
		Phone:       strings.TrimSpace(req.Phone),
		Role:        req.Role,
		AadhaarHash: strings.TrimSpace(req.AadhaarHash),
		KycVerified: false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	record := &Record{
		User:         user,
		PasswordHash: hashPassword(req.Password),
	}
	if err := s.repo.Create(ctx, record); err != nil {
		if errors.Is(err, ErrDuplicateEmail) {
			return nil, status.Error(codes.AlreadyExists, "email is already registered")
		}
		return nil, status.Errorf(codes.Internal, "create user: %v", err)
	}

	return &rentrelaypb.RegisterResponse{
		User:  cloneUser(user),
		Token: newToken(user.UserId),
	}, nil
}

func (s *Service) Login(ctx context.Context, req *rentrelaypb.LoginRequest) (*rentrelaypb.LoginResponse, error) {
	record, err := s.repo.FindByEmail(ctx, normalizeEmail(req.Email))
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, status.Error(codes.Unauthenticated, "invalid email or password")
		}
		return nil, status.Errorf(codes.Internal, "find user by email: %v", err)
	}

	if record.PasswordHash != hashPassword(req.Password) {
		return nil, status.Error(codes.Unauthenticated, "invalid email or password")
	}

	return &rentrelaypb.LoginResponse{
		Token:        newToken(record.User.UserId),
		RefreshToken: newToken("refresh-" + record.User.UserId),
		User:         cloneUser(record.User),
	}, nil
}

func (s *Service) GetUser(ctx context.Context, req *rentrelaypb.GetUserRequest) (*rentrelaypb.User, error) {
	record, err := s.repo.FindByID(ctx, req.UserId)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "find user by id: %v", err)
	}
	return cloneUser(record.User), nil
}

func (s *Service) UpdateKYC(ctx context.Context, req *rentrelaypb.UpdateKYCRequest) (*rentrelaypb.User, error) {
	record, err := s.repo.UpdateKYC(ctx, req.UserId, strings.TrimSpace(req.AadhaarHash), req.Verified)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "update kyc: %v", err)
	}
	return cloneUser(record.User), nil
}

func (s *Service) RefreshToken(ctx context.Context, req *rentrelaypb.LoginResponse) (*rentrelaypb.LoginResponse, error) {
	if strings.TrimSpace(req.RefreshToken) == "" || req.User == nil {
		return nil, status.Error(codes.InvalidArgument, "refresh token and user are required")
	}

	return &rentrelaypb.LoginResponse{
		Token:        newToken(req.User.UserId),
		RefreshToken: req.RefreshToken,
		User:         cloneUser(req.User),
	}, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func hashPassword(password string) string {
	sum := sha256.Sum256([]byte(password))
	return hex.EncodeToString(sum[:])
}

func newID(prefix string) string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
	}
	return fmt.Sprintf("%s-%s", prefix, hex.EncodeToString(b[:]))
}

func newToken(subject string) string {
	return fmt.Sprintf("dev-token-%s-%d", subject, time.Now().Unix())
}

func cloneRecord(record *Record) *Record {
	if record == nil {
		return nil
	}
	return &Record{
		User:         cloneUser(record.User),
		PasswordHash: record.PasswordHash,
	}
}

func cloneUser(user *rentrelaypb.User) *rentrelaypb.User {
	if user == nil {
		return nil
	}
	cp := *user
	return &cp
}
