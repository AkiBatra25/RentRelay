package property

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
)

type Service struct {
	rentrelaypb.UnimplementedPropertyServiceServer

	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func NewInMemoryService() *Service {
	return NewService(NewMemoryRepository())
}

func (s *Service) RegisterProperty(ctx context.Context, req *rentrelaypb.RegisterPropertyRequest) (*rentrelaypb.Property, error) {
	if strings.TrimSpace(req.LandlordId) == "" {
		return nil, status.Error(codes.InvalidArgument, "landlord_id is required")
	}
	if strings.TrimSpace(req.Title) == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}
	if strings.TrimSpace(req.City) == "" {
		return nil, status.Error(codes.InvalidArgument, "city is required")
	}
	if req.Bedrooms <= 0 {
		return nil, status.Error(codes.InvalidArgument, "bedrooms must be greater than zero")
	}
	if req.RentMonthly <= 0 {
		return nil, status.Error(codes.InvalidArgument, "rent_monthly must be greater than zero")
	}

	property := &rentrelaypb.Property{
		PropertyId:  newID("property"),
		LandlordId:  strings.TrimSpace(req.LandlordId),
		Title:       strings.TrimSpace(req.Title),
		Address:     strings.TrimSpace(req.Address),
		City:        strings.TrimSpace(req.City),
		Zone:        strings.TrimSpace(req.Zone),
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		Bedrooms:    req.Bedrooms,
		RentMonthly: req.RentMonthly,
		DepositAmt:  req.DepositAmt,
		Furnishing:  req.Furnishing,
		Amenities:   append([]string(nil), req.Amenities...),
		IsAvailable: true,
	}

	if err := s.repo.Create(ctx, property); err != nil {
		if errors.Is(err, ErrDuplicateProperty) {
			return nil, status.Error(codes.AlreadyExists, "property already exists")
		}
		return nil, status.Errorf(codes.Internal, "create property: %v", err)
	}

	return property, nil
}

func (s *Service) GetProperty(ctx context.Context, req *rentrelaypb.GetPropertyRequest) (*rentrelaypb.Property, error) {
	property, err := s.repo.FindByID(ctx, strings.TrimSpace(req.PropertyId))
	if err != nil {
		if errors.Is(err, ErrPropertyNotFound) {
			return nil, status.Error(codes.NotFound, "property not found")
		}
		return nil, status.Errorf(codes.Internal, "find property: %v", err)
	}

	return property, nil
}

func (s *Service) SearchProperties(ctx context.Context, req *rentrelaypb.SearchPropertiesRequest) (*rentrelaypb.SearchPropertiesResponse, error) {
	properties, err := s.repo.Search(ctx, SearchFilter{
		City:        strings.TrimSpace(req.City),
		Zone:        strings.TrimSpace(req.Zone),
		MinBedrooms: req.MinBedrooms,
		MaxRent:     req.MaxRent,
		Furnishing:  req.Furnishing,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "search properties: %v", err)
	}

	return &rentrelaypb.SearchPropertiesResponse{
		Properties: properties,
	}, nil
}

func (s *Service) UpdateAvailability(ctx context.Context, req *rentrelaypb.UpdateAvailabilityRequest) (*rentrelaypb.Property, error) {
	property, err := s.repo.UpdateAvailability(ctx, strings.TrimSpace(req.PropertyId), req.IsAvailable)
	if err != nil {
		if errors.Is(err, ErrPropertyNotFound) {
			return nil, status.Error(codes.NotFound, "property not found")
		}
		return nil, status.Errorf(codes.Internal, "update availability: %v", err)
	}

	return property, nil
}

func (s *Service) ListByLandlord(ctx context.Context, req *rentrelaypb.GetUserRequest) (*rentrelaypb.SearchPropertiesResponse, error) {
	properties, err := s.repo.ListByLandlord(ctx, strings.TrimSpace(req.UserId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list properties by landlord: %v", err)
	}

	return &rentrelaypb.SearchPropertiesResponse{
		Properties: properties,
	}, nil
}

func newID(prefix string) string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
	}
	return fmt.Sprintf("%s-%s", prefix, hex.EncodeToString(b[:]))
}
