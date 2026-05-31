package landlord

import (
	"context"
	"errors"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
)

var (
	ErrLeaseTermsNotFound = errors.New("lease terms not found")
)

type Repository interface {
	SaveLeaseTerms(ctx context.Context, terms *rentrelaypb.LeaseTerms) error
	FindLeaseTerms(ctx context.Context, landlordID string, propertyID string) (*rentrelaypb.LeaseTerms, error)
}
