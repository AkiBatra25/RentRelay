package user

import (
	"context"
	"errors"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
)

var (
	ErrDuplicateEmail = errors.New("email is already registered")
	ErrUserNotFound   = errors.New("user not found")
)

type Record struct {
	User         *rentrelaypb.User
	PasswordHash string
}

type Repository interface {
	Create(ctx context.Context, record *Record) error
	FindByEmail(ctx context.Context, email string) (*Record, error)
	FindByID(ctx context.Context, userID string) (*Record, error)
	UpdateKYC(ctx context.Context, userID string, aadhaarHash string, verified bool) (*Record, error)
}
