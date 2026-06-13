package document

import (
	"context"
	"testing"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
)

func TestUploadVerifyAndLockDocument(t *testing.T) {
	svc := NewInMemoryService()
	ctx := context.Background()

	uploaded, err := svc.UploadDocument(ctx, &rentrelaypb.UploadDocumentRequest{
		AgreementId: "agreement-1",
		DocType:     "AGREEMENT",
		UploadedBy:  "tenant-1",
		Filename:    "agreement.txt",
		Content:     []byte("signed agreement"),
	})
	if err != nil {
		t.Fatalf("UploadDocument() error = %v", err)
	}

	verified, err := svc.VerifyDocument(ctx, &rentrelaypb.VerifyDocumentRequest{
		DocumentId: uploaded.DocumentId,
		Sha256Hash: uploaded.Sha256Hash,
	})
	if err != nil || !verified.IsValid {
		t.Fatalf("VerifyDocument() valid = %v, error = %v", verified.GetIsValid(), err)
	}

	if _, err := svc.LockDocuments(ctx, &rentrelaypb.LockDocumentsRequest{AgreementId: "agreement-1"}); err != nil {
		t.Fatalf("LockDocuments() error = %v", err)
	}
	list, err := svc.ListByAgreement(ctx, &rentrelaypb.AgreementActionRequest{AgreementId: "agreement-1"})
	if err != nil {
		t.Fatalf("ListByAgreement() error = %v", err)
	}
	if len(list.Documents) != 1 || !list.Documents[0].Locked {
		t.Fatalf("locked documents = %#v", list.Documents)
	}
}
