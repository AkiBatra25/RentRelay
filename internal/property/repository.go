package property

import (
	"context"
	"errors"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
)

var (
	ErrDuplicateProperty = errors.New("property already exists")
	ErrPropertyNotFound  = errors.New("property not found")
)

type SearchFilter struct {
	City        string
	Zone        string
	MinBedrooms int32
	MaxRent     float64
	Furnishing  rentrelaypb.FurnishingType
}

type Repository interface {
	Create(ctx context.Context, property *rentrelaypb.Property) error
	FindByID(ctx context.Context, propertyID string) (*rentrelaypb.Property, error)
	Search(ctx context.Context, filter SearchFilter) ([]*rentrelaypb.Property, error)
	UpdateAvailability(ctx context.Context, propertyID string, isAvailable bool) (*rentrelaypb.Property, error)
	ListByLandlord(ctx context.Context, landlordID string) ([]*rentrelaypb.Property, error)
}