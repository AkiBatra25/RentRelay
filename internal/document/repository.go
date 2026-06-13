package document

import (
	"context"
	"errors"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
)

var ErrDocumentNotFound = errors.New("document not found")

type Record struct {
	Document *rentrelaypb.Document
	Filename string
	Content  []byte
}

type Repository interface {
	Create(ctx context.Context, record *Record) error
	FindByID(ctx context.Context, documentID string) (*Record, error)
	FindFirstByAgreement(ctx context.Context, agreementID string) (*Record, error)
	ListByAgreement(ctx context.Context, agreementID string) ([]*rentrelaypb.Document, error)
	SetLockedByAgreement(ctx context.Context, agreementID string, locked bool) error
	Close(ctx context.Context) error
}
