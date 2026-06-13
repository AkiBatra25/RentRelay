package agreement

import (
	"context"
	"testing"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAgreementHappyPathToActive(t *testing.T) {
	svc := NewInMemoryService()
	ctx := context.Background()

	created, err := svc.CreateAgreement(ctx, &rentrelaypb.CreateAgreementRequest{
		TenantId:      "tenant-1",
		LandlordId:    "landlord-1",
		PropertyId:    "property-1",
		MonthlyRent:   25000,
		DepositAmount: 75000,
		LeaseMonths:   11,
		NoticeDays:    30,
	})
	if err != nil {
		t.Fatalf("CreateAgreement() error = %v", err)
	}

	firstSign, err := svc.SignAgreement(ctx, &rentrelaypb.SignAgreementRequest{
		AgreementId:   created.AgreementId,
		SignerId:      "tenant-1",
		SignatureHash: "tenant-signature",
	})
	if err != nil {
		t.Fatalf("SignAgreement() tenant error = %v", err)
	}
	if firstSign.State != rentrelaypb.AgreementState_CREATED {
		t.Fatalf("state after one signature = %s, want CREATED", firstSign.State)
	}

	signed, err := svc.SignAgreement(ctx, &rentrelaypb.SignAgreementRequest{
		AgreementId:   created.AgreementId,
		SignerId:      "landlord-1",
		SignatureHash: "landlord-signature",
	})
	if err != nil {
		t.Fatalf("SignAgreement() landlord error = %v", err)
	}
	if signed.State != rentrelaypb.AgreementState_SIGNED {
		t.Fatalf("state after both signatures = %s, want SIGNED", signed.State)
	}

	escrow, err := svc.HoldEscrow(ctx, &rentrelaypb.AgreementActionRequest{
		AgreementId: created.AgreementId,
		ActorId:     "tenant-1",
	})
	if err != nil {
		t.Fatalf("HoldEscrow() error = %v", err)
	}
	if escrow.DepositHeld != 75000 {
		t.Fatalf("deposit_held = %.2f, want 75000", escrow.DepositHeld)
	}

	active, err := svc.StartLease(ctx, &rentrelaypb.AgreementActionRequest{
		AgreementId: created.AgreementId,
		ActorId:     "landlord-1",
	})
	if err != nil {
		t.Fatalf("StartLease() error = %v", err)
	}
	if active.State != rentrelaypb.AgreementState_ACTIVE {
		t.Fatalf("state = %s, want ACTIVE", active.State)
	}
}

func TestAgreementRejectsInvalidTransition(t *testing.T) {
	svc := NewInMemoryService()
	created, err := svc.CreateAgreement(context.Background(), &rentrelaypb.CreateAgreementRequest{
		TenantId:      "tenant-1",
		LandlordId:    "landlord-1",
		PropertyId:    "property-1",
		MonthlyRent:   25000,
		DepositAmount: 75000,
		LeaseMonths:   11,
		NoticeDays:    30,
	})
	if err != nil {
		t.Fatalf("CreateAgreement() error = %v", err)
	}

	_, err = svc.StartLease(context.Background(), &rentrelaypb.AgreementActionRequest{
		AgreementId: created.AgreementId,
		ActorId:     "landlord-1",
	})
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("StartLease() code = %v, want %v", status.Code(err), codes.FailedPrecondition)
	}
}
