package document

import (
	"context"
	"sync"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
)

type MemoryRepository struct {
	mu          sync.RWMutex
	recordsByID map[string]*Record
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{recordsByID: make(map[string]*Record)}
}

func (r *MemoryRepository) Create(ctx context.Context, record *Record) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.recordsByID[record.Document.DocumentId] = cloneRecord(record)
	return nil
}

func (r *MemoryRepository) FindByID(ctx context.Context, documentID string) (*Record, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	record, ok := r.recordsByID[documentID]
	if !ok {
		return nil, ErrDocumentNotFound
	}
	return cloneRecord(record), nil
}

func (r *MemoryRepository) FindFirstByAgreement(ctx context.Context, agreementID string) (*Record, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, record := range r.recordsByID {
		if record.Document.AgreementId == agreementID {
			return cloneRecord(record), nil
		}
	}
	return nil, ErrDocumentNotFound
}

func (r *MemoryRepository) ListByAgreement(ctx context.Context, agreementID string) ([]*rentrelaypb.Document, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var documents []*rentrelaypb.Document
	for _, record := range r.recordsByID {
		if record.Document.AgreementId == agreementID {
			documents = append(documents, cloneDocument(record.Document))
		}
	}
	return documents, nil
}

func (r *MemoryRepository) SetLockedByAgreement(ctx context.Context, agreementID string, locked bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, record := range r.recordsByID {
		if record.Document.AgreementId == agreementID {
			record.Document.Locked = locked
		}
	}
	return nil
}

func (r *MemoryRepository) Close(context.Context) error { return nil }

func cloneRecord(record *Record) *Record {
	if record == nil {
		return nil
	}
	return &Record{
		Document: cloneDocument(record.Document),
		Filename: record.Filename,
		Content:  append([]byte(nil), record.Content...),
	}
}

func cloneDocument(document *rentrelaypb.Document) *rentrelaypb.Document {
	if document == nil {
		return nil
	}
	cp := *document
	return &cp
}
