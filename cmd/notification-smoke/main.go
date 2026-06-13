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
	conn, err := grpc.NewClient(envOrDefault("NOTIFICATION_SERVICE_ADDR", "localhost:50057"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client := rentrelaypb.NewNotificationServiceClient(conn)
	userID := fmt.Sprintf("notification-user-%d", time.Now().UnixNano())
	item, err := client.Send(ctx, &rentrelaypb.SendNotificationRequest{
		UserId: userID, AgreementId: "agreement-1",
		Event:    rentrelaypb.NotificationEvent_NOTIFICATION_AGREEMENT_CREATED,
		Channels: []rentrelaypb.NotificationChannel{rentrelaypb.NotificationChannel_EMAIL, rentrelaypb.NotificationChannel_PUSH},
	})
	if err != nil {
		log.Fatal(err)
	}
	history, err := client.GetHistory(ctx, &rentrelaypb.GetUserRequest{UserId: userID})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("notification_id=%s delivered=%v history_total=%d\n", item.NotificationId, item.Delivered, history.Total)
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
