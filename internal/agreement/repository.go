package agreement

import (
	"context"
	"errors"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
)

var ErrAgreementNotFound = errors.New("agreement not found")

type Repository interface {
	Create(ctx context.Context, agreement *rentrelaypb.Agreement) error
	FindByID(ctx context.Context, agreementID string) (*rentrelaypb.Agreement, error)
	Save(ctx context.Context, agreement *rentrelaypb.Agreement) error
}
