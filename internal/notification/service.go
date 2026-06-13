package notification

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	rentrelaypb.UnimplementedNotificationServiceServer
	repo Repository

	mu          sync.RWMutex
	subscribers map[string]map[chan *rentrelaypb.Notification]struct{}
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo, subscribers: make(map[string]map[chan *rentrelaypb.Notification]struct{})}
}

func NewInMemoryService() *Service { return NewService(NewMemoryRepository()) }

func (s *Service) Send(ctx context.Context, req *rentrelaypb.SendNotificationRequest) (*rentrelaypb.Notification, error) {
	if req == nil || strings.TrimSpace(req.UserId) == "" || req.Event == rentrelaypb.NotificationEvent_NOTIFICATION_EVENT_UNKNOWN {
		return nil, status.Error(codes.InvalidArgument, "user_id and event are required")
	}
	channels := req.Channels
	if len(channels) == 0 {
		channels = []rentrelaypb.NotificationChannel{rentrelaypb.NotificationChannel_PUSH}
	}

	var first *rentrelaypb.Notification
	for _, channel := range channels {
		now := timestamppb.Now()
		item := &rentrelaypb.Notification{
			NotificationId: newID("notification"), UserId: strings.TrimSpace(req.UserId),
			AgreementId: req.AgreementId, Event: req.Event, Channel: channel,
			Message: renderMessage(req.Event, req.TemplateVars), Delivered: true,
			CreatedAt: now, DeliveredAt: now,
		}
		if err := s.repo.Create(ctx, item); err != nil {
			return nil, status.Errorf(codes.Internal, "store notification: %v", err)
		}
		s.publish(item)
		if first == nil {
			first = item
		}
	}
	return cloneNotification(first), nil
}

func (s *Service) Broadcast(ctx context.Context, req *rentrelaypb.BroadcastRequest) (*emptypb.Empty, error) {
	if req == nil || len(req.UserIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_ids are required")
	}
	for _, userID := range req.UserIds {
		if _, err := s.Send(ctx, &rentrelaypb.SendNotificationRequest{
			UserId: userID, AgreementId: req.AgreementId, Event: req.Event,
			Channels: req.Channels, TemplateVars: req.TemplateVars,
		}); err != nil {
			return nil, err
		}
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) GetHistory(ctx context.Context, req *rentrelaypb.GetUserRequest) (*rentrelaypb.NotificationList, error) {
	if req == nil || strings.TrimSpace(req.UserId) == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	items, err := s.repo.ListByUser(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list notifications: %v", err)
	}
	return &rentrelaypb.NotificationList{Notifications: items, Total: int32(len(items))}, nil
}

func (s *Service) Subscribe(req *rentrelaypb.GetUserRequest, stream grpc.ServerStreamingServer[rentrelaypb.Notification]) error {
	if req == nil || strings.TrimSpace(req.UserId) == "" {
		return status.Error(codes.InvalidArgument, "user_id is required")
	}
	ch := make(chan *rentrelaypb.Notification, 16)
	s.addSubscriber(req.UserId, ch)
	defer s.removeSubscriber(req.UserId, ch)

	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case item := <-ch:
			if err := stream.Send(item); err != nil {
				return err
			}
		}
	}
}

func (s *Service) publish(item *rentrelaypb.Notification) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for ch := range s.subscribers[item.UserId] {
		select {
		case ch <- cloneNotification(item):
		default:
		}
	}
}

func (s *Service) addSubscriber(userID string, ch chan *rentrelaypb.Notification) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.subscribers[userID] == nil {
		s.subscribers[userID] = make(map[chan *rentrelaypb.Notification]struct{})
	}
	s.subscribers[userID][ch] = struct{}{}
}

func (s *Service) removeSubscriber(userID string, ch chan *rentrelaypb.Notification) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.subscribers[userID], ch)
	close(ch)
}

func renderMessage(event rentrelaypb.NotificationEvent, vars map[string]string) string {
	message := strings.ReplaceAll(event.String(), "NOTIFICATION_", "")
	if len(vars) == 0 {
		return message
	}
	return fmt.Sprintf("%s %v", message, vars)
}

func newID(prefix string) string {
	var value [8]byte
	if _, err := rand.Read(value[:]); err != nil {
		return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
	}
	return prefix + "-" + hex.EncodeToString(value[:])
}
