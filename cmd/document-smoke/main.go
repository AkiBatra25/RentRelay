package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient(envOrDefault("DOCUMENT_SERVICE_ADDR", "localhost:50058"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client := rentrelaypb.NewDocumentServiceClient(conn)
	agreementID := fmt.Sprintf("agreement-document-%d", time.Now().UnixNano())
	doc, err := client.UploadDocument(ctx, &rentrelaypb.UploadDocumentRequest{
		AgreementId: agreementID, DocType: "AGREEMENT", UploadedBy: "tenant-1",
		Filename: "agreement.txt", Content: []byte("signed rental agreement"),
	})
	if err != nil {
		log.Fatal(err)
	}
	verify, err := client.VerifyDocument(ctx, &rentrelaypb.VerifyDocumentRequest{DocumentId: doc.DocumentId, Sha256Hash: doc.Sha256Hash})
	if err != nil {
		log.Fatal(err)
	}
	_, _ = client.LockDocuments(ctx, &rentrelaypb.LockDocumentsRequest{AgreementId: agreementID})
	list, err := client.ListByAgreement(ctx, &rentrelaypb.AgreementActionRequest{AgreementId: agreementID})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("document_id=%s hash_valid=%v locked=%v\n", doc.DocumentId, verify.IsValid, list.Documents[0].Locked)
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
