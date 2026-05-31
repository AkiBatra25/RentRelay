package landlord

import (
	"context"
	"sync"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
)

type MemoryRepository struct {
	mu        sync.RWMutex
	termsByID map[string]*rentrelaypb.LeaseTerms
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		termsByID: make(map[string]*rentrelaypb.LeaseTerms),
	}
}

func (r *MemoryRepository) SaveLeaseTerms(ctx context.Context, terms *rentrelaypb.LeaseTerms) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.termsByID[leaseTermsKey(terms.LandlordId, terms.PropertyId)] = cloneLeaseTerms(terms)
	return nil
}

func (r *MemoryRepository) FindLeaseTerms(ctx context.Context, landlordID string, propertyID string) (*rentrelaypb.LeaseTerms, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	terms, exists := r.termsByID[leaseTermsKey(landlordID, propertyID)]
	if !exists {
		return nil, ErrLeaseTermsNotFound
	}

	return cloneLeaseTerms(terms), nil
}

func leaseTermsKey(landlordID string, propertyID string) string {
	return landlordID + ":" + propertyID
}

func cloneLeaseTerms(terms *rentrelaypb.LeaseTerms) *rentrelaypb.LeaseTerms {
	if terms == nil {
		return nil
	}

	cp := *terms
	cp.AllowedTypes = append([]string(nil), terms.AllowedTypes...)
	return &cp
}
