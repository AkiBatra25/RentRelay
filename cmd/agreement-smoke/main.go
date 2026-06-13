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
	addr := os.Getenv("AGREEMENT_SERVICE_ADDR")
	if addr == "" {
		addr = "localhost:50055"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("create grpc client: %v", err)
	}
	defer conn.Close()

	client := rentrelaypb.NewAgreementServiceClient(conn)
	suffix := time.Now().UnixNano()
	tenantID := fmt.Sprintf("tenant-%d", suffix)
	landlordID := fmt.Sprintf("landlord-%d", suffix)

	created, err := client.CreateAgreement(ctx, &rentrelaypb.CreateAgreementRequest{
		TenantId:      tenantID,
		LandlordId:    landlordID,
		PropertyId:    fmt.Sprintf("property-%d", suffix),
		MonthlyRent:   25000,
		DepositAmount: 75000,
		LeaseMonths:   11,
		NoticeDays:    30,
	})
	if err != nil {
		log.Fatalf("create agreement: %v", err)
	}

	if _, err := client.SignAgreement(ctx, &rentrelaypb.SignAgreementRequest{
		AgreementId:   created.AgreementId,
		SignerId:      tenantID,
		SignatureHash: "tenant-signature-hash",
	}); err != nil {
		log.Fatalf("tenant sign agreement: %v", err)
	}

	signed, err := client.SignAgreement(ctx, &rentrelaypb.SignAgreementRequest{
		AgreementId:   created.AgreementId,
		SignerId:      landlordID,
		SignatureHash: "landlord-signature-hash",
	})
	if err != nil {
		log.Fatalf("landlord sign agreement: %v", err)
	}

	escrow, err := client.HoldEscrow(ctx, &rentrelaypb.AgreementActionRequest{
		AgreementId: created.AgreementId,
		ActorId:     tenantID,
	})
	if err != nil {
		log.Fatalf("hold escrow: %v", err)
	}

	active, err := client.StartLease(ctx, &rentrelaypb.AgreementActionRequest{
		AgreementId: created.AgreementId,
		ActorId:     landlordID,
	})
	if err != nil {
		log.Fatalf("start lease: %v", err)
	}

	fmt.Printf("created agreement_id=%s state=%s\n", created.AgreementId, created.State)
	fmt.Printf("signed state=%s signatures=%d\n", signed.State, len(signed.Signatures))
	fmt.Printf("escrow state=%s deposit_held=%.2f\n", escrow.State, escrow.DepositHeld)
	fmt.Printf("lease state=%s replica_version=%d\n", active.State, active.ReplicaVersion)
}
