package tenant

import (
	"context"
	"errors"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
)

var ErrRentalRequestNotFound = errors.New("rental request not found")

type Repository interface {
	CreateRentalRequest(ctx context.Context, request *rentrelaypb.RentalRequest) error
	FindRentalRequestByTenant(ctx context.Context, tenantID string) (*rentrelaypb.RentalRequest, error)
}
