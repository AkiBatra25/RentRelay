package notification

import (
	"context"
	"sync"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
)

type MemoryRepository struct {
	mu            sync.RWMutex
	notifications []*rentrelaypb.Notification
}

func NewMemoryRepository() *MemoryRepository { return &MemoryRepository{} }

func (r *MemoryRepository) Create(ctx context.Context, notification *rentrelaypb.Notification) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.notifications = append(r.notifications, cloneNotification(notification))
	return nil
}

func (r *MemoryRepository) ListByUser(ctx context.Context, userID string) ([]*rentrelaypb.Notification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var results []*rentrelaypb.Notification
	for _, item := range r.notifications {
		if item.UserId == userID {
			results = append(results, cloneNotification(item))
		}
	}
	return results, nil
}

func (r *MemoryRepository) Close(context.Context) error { return nil }

func cloneNotification(value *rentrelaypb.Notification) *rentrelaypb.Notification {
	if value == nil {
		return nil
	}
	cp := *value
	return &cp
}
