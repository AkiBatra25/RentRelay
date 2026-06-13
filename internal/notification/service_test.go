package notification

import (
	"context"
	"testing"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
)

func TestSendAndGetHistory(t *testing.T) {
	svc := NewInMemoryService()
	ctx := context.Background()

	if _, err := svc.Send(ctx, &rentrelaypb.SendNotificationRequest{
		UserId: "user-1", AgreementId: "agreement-1",
		Event:    rentrelaypb.NotificationEvent_NOTIFICATION_AGREEMENT_CREATED,
		Channels: []rentrelaypb.NotificationChannel{rentrelaypb.NotificationChannel_EMAIL, rentrelaypb.NotificationChannel_PUSH},
	}); err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	history, err := svc.GetHistory(ctx, &rentrelaypb.GetUserRequest{UserId: "user-1"})
	if err != nil {
		t.Fatalf("GetHistory() error = %v", err)
	}
	if history.Total != 2 {
		t.Fatalf("history total = %d, want 2", history.Total)
	}
}
