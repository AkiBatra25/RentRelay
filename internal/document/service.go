package document

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	rentrelaypb.UnimplementedDocumentServiceServer
	repo Repository
}

func NewService(repo Repository) *Service { return &Service{repo: repo} }
func NewInMemoryService() *Service        { return NewService(NewMemoryRepository()) }

func (s *Service) UploadDocument(ctx context.Context, req *rentrelaypb.UploadDocumentRequest) (*rentrelaypb.Document, error) {
	if req == nil || strings.TrimSpace(req.AgreementId) == "" || strings.TrimSpace(req.DocType) == "" ||
		strings.TrimSpace(req.UploadedBy) == "" || len(req.Content) == 0 {
		return nil, status.Error(codes.InvalidArgument, "agreement_id, doc_type, uploaded_by, and content are required")
	}

	id := newID("document")
	hash := sha256.Sum256(req.Content)
	document := &rentrelaypb.Document{
		DocumentId:  id,
		AgreementId: strings.TrimSpace(req.AgreementId),
		DocType:     strings.TrimSpace(req.DocType),
		StorageKey:  "documents/" + id + "/" + strings.TrimSpace(req.Filename),
		Sha256Hash:  hex.EncodeToString(hash[:]),
		SizeBytes:   int64(len(req.Content)),
		UploadedBy:  strings.TrimSpace(req.UploadedBy),
		UploadedAt:  timestamppb.Now(),
	}
	if err := s.repo.Create(ctx, &Record{Document: document, Filename: req.Filename, Content: req.Content}); err != nil {
		return nil, status.Errorf(codes.Internal, "store document: %v", err)
	}
	return cloneDocument(document), nil
}

func (s *Service) GetDocument(ctx context.Context, req *rentrelaypb.AgreementActionRequest) (*rentrelaypb.Document, error) {
	if req == nil || strings.TrimSpace(req.AgreementId) == "" {
		return nil, status.Error(codes.InvalidArgument, "agreement_id is required")
	}
	record, err := s.repo.FindFirstByAgreement(ctx, req.AgreementId)
	if errors.Is(err, ErrDocumentNotFound) {
		return nil, status.Error(codes.NotFound, "document not found")
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get document: %v", err)
	}
	return record.Document, nil
}

func (s *Service) VerifyDocument(ctx context.Context, req *rentrelaypb.VerifyDocumentRequest) (*rentrelaypb.VerifyDocumentResponse, error) {
	if req == nil || strings.TrimSpace(req.DocumentId) == "" || strings.TrimSpace(req.Sha256Hash) == "" {
		return nil, status.Error(codes.InvalidArgument, "document_id and sha256_hash are required")
	}
	record, err := s.repo.FindByID(ctx, req.DocumentId)
	if errors.Is(err, ErrDocumentNotFound) {
		return nil, status.Error(codes.NotFound, "document not found")
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "verify document: %v", err)
	}
	return &rentrelaypb.VerifyDocumentResponse{
		IsValid:    strings.EqualFold(record.Document.Sha256Hash, req.Sha256Hash),
		DocumentId: record.Document.DocumentId,
		StoredHash: record.Document.Sha256Hash,
	}, nil
}

func (s *Service) LockDocuments(ctx context.Context, req *rentrelaypb.LockDocumentsRequest) (*emptypb.Empty, error) {
	return s.setLocked(ctx, req, true)
}

func (s *Service) UnlockDocuments(ctx context.Context, req *rentrelaypb.LockDocumentsRequest) (*emptypb.Empty, error) {
	return s.setLocked(ctx, req, false)
}

func (s *Service) ListByAgreement(ctx context.Context, req *rentrelaypb.AgreementActionRequest) (*rentrelaypb.DocumentList, error) {
	if req == nil || strings.TrimSpace(req.AgreementId) == "" {
		return nil, status.Error(codes.InvalidArgument, "agreement_id is required")
	}
	documents, err := s.repo.ListByAgreement(ctx, req.AgreementId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list documents: %v", err)
	}
	return &rentrelaypb.DocumentList{Documents: documents}, nil
}

func (s *Service) setLocked(ctx context.Context, req *rentrelaypb.LockDocumentsRequest, locked bool) (*emptypb.Empty, error) {
	if req == nil || strings.TrimSpace(req.AgreementId) == "" {
		return nil, status.Error(codes.InvalidArgument, "agreement_id is required")
	}
	if err := s.repo.SetLockedByAgreement(ctx, req.AgreementId, locked); err != nil {
		return nil, status.Errorf(codes.Internal, "set document lock: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func newID(prefix string) string {
	var value [8]byte
	if _, err := rand.Read(value[:]); err != nil {
		return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
	}
	return prefix + "-" + hex.EncodeToString(value[:])
}
