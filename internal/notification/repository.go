package notification

import (
	"context"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
)

type Repository interface {
	Create(ctx context.Context, notification *rentrelaypb.Notification) error
	ListByUser(ctx context.Context, userID string) ([]*rentrelaypb.Notification, error)
	Close(ctx context.Context) error
}
