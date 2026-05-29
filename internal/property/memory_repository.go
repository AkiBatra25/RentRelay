package property

import (
	"context"
	"strings"
	"sync"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
)

type MemoryRepository struct {
	mu             sync.RWMutex
	propertiesByID map[string]*rentrelaypb.Property
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		propertiesByID: make(map[string]*rentrelaypb.Property),
	}
}

func (r *MemoryRepository) Create(ctx context.Context, property *rentrelaypb.Property) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.propertiesByID[property.PropertyId]; exists {
		return ErrDuplicateProperty
	}

	r.propertiesByID[property.PropertyId] = cloneProperty(property)
	return nil
}

func (r *MemoryRepository) FindByID(ctx context.Context, propertyID string) (*rentrelaypb.Property, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	property, exists := r.propertiesByID[propertyID]
	if !exists {
		return nil, ErrPropertyNotFound
	}

	return cloneProperty(property), nil
}

func (r *MemoryRepository) Search(ctx context.Context, filter SearchFilter) ([]*rentrelaypb.Property, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []*rentrelaypb.Property

	for _, property := range r.propertiesByID {
		if !property.IsAvailable {
			continue
		}
		if filter.City != "" && !strings.EqualFold(property.City, filter.City) {
			continue
		}
		if filter.Zone != "" && !strings.EqualFold(property.Zone, filter.Zone) {
			continue
		}
		if filter.MinBedrooms > 0 && property.Bedrooms < filter.MinBedrooms {
			continue
		}
		if filter.MaxRent > 0 && property.RentMonthly > filter.MaxRent {
			continue
		}
		if filter.Furnishing != rentrelaypb.FurnishingType_FURNISHING_UNKNOWN && property.Furnishing != filter.Furnishing {
			continue
		}

		results = append(results, cloneProperty(property))
	}

	return results, nil
}

func (r *MemoryRepository) UpdateAvailability(ctx context.Context, propertyID string, isAvailable bool) (*rentrelaypb.Property, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	property, exists := r.propertiesByID[propertyID]
	if !exists {
		return nil, ErrPropertyNotFound
	}

	property.IsAvailable = isAvailable
	return cloneProperty(property), nil
}

func (r *MemoryRepository) ListByLandlord(ctx context.Context, landlordID string) ([]*rentrelaypb.Property, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []*rentrelaypb.Property

	for _, property := range r.propertiesByID {
		if property.LandlordId == landlordID {
			results = append(results, cloneProperty(property))
		}
	}

	return results, nil
}

func cloneProperty(property *rentrelaypb.Property) *rentrelaypb.Property {
	if property == nil {
		return nil
	}

	cp := *property
	cp.Amenities = append([]string(nil), property.Amenities...)
	return &cp
}
