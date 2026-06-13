package agreement

import (
	"context"
	"errors"
	"fmt"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
)

type AgreementReplicator interface {
	Store(ctx context.Context, agreement *rentrelaypb.Agreement) error
	Load(ctx context.Context, agreementID string) (*rentrelaypb.Agreement, error)
}

type ReplicatedRepository struct {
	primary    Repository
	replicator AgreementReplicator
}

func NewReplicatedRepository(primary Repository, replicator AgreementReplicator) *ReplicatedRepository {
	return &ReplicatedRepository{primary: primary, replicator: replicator}
}

func (r *ReplicatedRepository) Create(ctx context.Context, agreement *rentrelaypb.Agreement) error {
	if err := r.replicator.Store(ctx, agreement); err != nil {
		return fmt.Errorf("replicate agreement: %w", err)
	}
	if err := r.primary.Create(ctx, agreement); err != nil {
		return fmt.Errorf("create primary agreement: %w", err)
	}
	return nil
}

func (r *ReplicatedRepository) FindByID(ctx context.Context, agreementID string) (*rentrelaypb.Agreement, error) {
	agreement, err := r.primary.FindByID(ctx, agreementID)
	if err == nil {
		return agreement, nil
	}
	if !errors.Is(err, ErrAgreementNotFound) {
		return nil, err
	}

	agreement, err = r.replicator.Load(ctx, agreementID)
	if err != nil {
		return nil, err
	}
	return agreement, nil
}

func (r *ReplicatedRepository) Save(ctx context.Context, agreement *rentrelaypb.Agreement) error {
	if err := r.replicator.Store(ctx, agreement); err != nil {
		return fmt.Errorf("replicate agreement: %w", err)
	}
	if err := r.primary.Save(ctx, agreement); err != nil {
		return fmt.Errorf("save primary agreement: %w", err)
	}
	return nil
}
