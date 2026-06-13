package tenant

import (
	"context"
	"sync"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
)

type MemoryRepository struct {
	mu                 sync.RWMutex
	requestsByTenantID map[string]*rentrelaypb.RentalRequest
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		requestsByTenantID: make(map[string]*rentrelaypb.RentalRequest),
	}
}

func (r *MemoryRepository) CreateRentalRequest(ctx context.Context, request *rentrelaypb.RentalRequest) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.requestsByTenantID[request.TenantId] = cloneRentalRequest(request)
	return nil
}

func (r *MemoryRepository) FindRentalRequestByTenant(ctx context.Context, tenantID string) (*rentrelaypb.RentalRequest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	request, exists := r.requestsByTenantID[tenantID]
	if !exists {
		return nil, ErrRentalRequestNotFound
	}

	return cloneRentalRequest(request), nil
}

func cloneRentalRequest(request *rentrelaypb.RentalRequest) *rentrelaypb.RentalRequest {
	if request == nil {
		return nil
	}
	cp := *request
	return &cp
}
