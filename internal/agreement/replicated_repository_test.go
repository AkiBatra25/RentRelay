package agreement

import (
	"context"
	"errors"
	"testing"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
)

type fakeReplicator struct {
	stored *rentrelaypb.Agreement
	load   *rentrelaypb.Agreement
	err    error
}

func (f *fakeReplicator) Store(_ context.Context, agreement *rentrelaypb.Agreement) error {
	if f.err != nil {
		return f.err
	}
	f.stored = cloneAgreement(agreement)
	f.stored.WorkerNode = "storage-worker-2"
	agreement.WorkerNode = f.stored.WorkerNode
	return nil
}

func (f *fakeReplicator) Load(context.Context, string) (*rentrelaypb.Agreement, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.load == nil {
		return nil, ErrAgreementNotFound
	}
	return cloneAgreement(f.load), nil
}

func TestReplicatedRepositoryStoresAgreementInBothRepositories(t *testing.T) {
	primary := NewMemoryRepository()
	replicator := &fakeReplicator{}
	repo := NewReplicatedRepository(primary, replicator)
	agreement := &rentrelaypb.Agreement{AgreementId: "agreement-1", ReplicaVersion: 1}

	if err := repo.Create(context.Background(), agreement); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if replicator.stored == nil {
		t.Fatal("agreement was not sent to replicated storage")
	}

	stored, err := primary.FindByID(context.Background(), agreement.AgreementId)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if stored.WorkerNode != "storage-worker-2" {
		t.Fatalf("worker_node = %q, want storage-worker-2", stored.WorkerNode)
	}
}

func TestReplicatedRepositoryFallsBackToStorage(t *testing.T) {
	recovered := &rentrelaypb.Agreement{AgreementId: "agreement-recovered", ReplicaVersion: 4}
	repo := NewReplicatedRepository(NewMemoryRepository(), &fakeReplicator{load: recovered})

	got, err := repo.FindByID(context.Background(), recovered.AgreementId)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if got.AgreementId != recovered.AgreementId {
		t.Fatalf("agreement_id = %q, want %q", got.AgreementId, recovered.AgreementId)
	}
}

func TestReplicatedRepositoryRequiresSuccessfulReplication(t *testing.T) {
	primary := NewMemoryRepository()
	repo := NewReplicatedRepository(primary, &fakeReplicator{err: errors.New("quorum unavailable")})
	agreement := &rentrelaypb.Agreement{AgreementId: "agreement-1"}

	if err := repo.Create(context.Background(), agreement); err == nil {
		t.Fatal("Create() error = nil, want replication failure")
	}
	if _, err := primary.FindByID(context.Background(), agreement.AgreementId); !errors.Is(err, ErrAgreementNotFound) {
		t.Fatalf("primary repository error = %v, want ErrAgreementNotFound", err)
	}
}
