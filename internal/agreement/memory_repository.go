package agreement

import (
	"context"
	"sync"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
)

type MemoryRepository struct {
	mu             sync.RWMutex
	agreementsByID map[string]*rentrelaypb.Agreement
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		agreementsByID: make(map[string]*rentrelaypb.Agreement),
	}
}

func (r *MemoryRepository) Create(ctx context.Context, agreement *rentrelaypb.Agreement) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.agreementsByID[agreement.AgreementId] = cloneAgreement(agreement)
	return nil
}

func (r *MemoryRepository) FindByID(ctx context.Context, agreementID string) (*rentrelaypb.Agreement, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agreement, exists := r.agreementsByID[agreementID]
	if !exists {
		return nil, ErrAgreementNotFound
	}
	return cloneAgreement(agreement), nil
}

func (r *MemoryRepository) Save(ctx context.Context, agreement *rentrelaypb.Agreement) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.agreementsByID[agreement.AgreementId]; !exists {
		return ErrAgreementNotFound
	}
	r.agreementsByID[agreement.AgreementId] = cloneAgreement(agreement)
	return nil
}

func cloneAgreement(agreement *rentrelaypb.Agreement) *rentrelaypb.Agreement {
	if agreement == nil {
		return nil
	}
	cp := *agreement
	cp.Signatures = append([]string(nil), agreement.Signatures...)
	return &cp
}
