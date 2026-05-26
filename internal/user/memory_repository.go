package user

import (
	"context"
	"sync"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type MemoryRepository struct {
	mu          sync.RWMutex
	usersByID   map[string]*Record
	usersByMail map[string]*Record
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		usersByID:   make(map[string]*Record),
		usersByMail: make(map[string]*Record),
	}
}

func (r *MemoryRepository) Create(ctx context.Context, record *Record) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.usersByMail[record.User.Email]; exists {
		return ErrDuplicateEmail
	}

	stored := cloneRecord(record)
	r.usersByID[record.User.UserId] = stored
	r.usersByMail[record.User.Email] = stored
	return nil
}

func (r *MemoryRepository) FindByEmail(ctx context.Context, email string) (*Record, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	record, exists := r.usersByMail[email]
	if !exists {
		return nil, ErrUserNotFound
	}
	return cloneRecord(record), nil
}

func (r *MemoryRepository) FindByID(ctx context.Context, userID string) (*Record, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	record, exists := r.usersByID[userID]
	if !exists {
		return nil, ErrUserNotFound
	}
	return cloneRecord(record), nil
}

func (r *MemoryRepository) UpdateKYC(ctx context.Context, userID string, aadhaarHash string, verified bool) (*Record, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	record, exists := r.usersByID[userID]
	if !exists {
		return nil, ErrUserNotFound
	}

	record.User.AadhaarHash = aadhaarHash
	record.User.KycVerified = verified
	record.User.UpdatedAt = timestamppb.Now()
	return cloneRecord(record), nil
}
