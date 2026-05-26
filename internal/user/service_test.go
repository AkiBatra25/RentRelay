package user

import (
	"context"
	"testing"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRegisterAndLogin(t *testing.T) {
	svc := NewInMemoryService()

	registerResp, err := svc.Register(context.Background(), &rentrelaypb.RegisterRequest{
		Name:     "Priya Reddy",
		Email:    "Priya@Test.com",
		Phone:    "9123456789",
		Password: "pass123",
		Role:     rentrelaypb.UserRole_ROLE_TENANT,
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if registerResp.User == nil {
		t.Fatal("Register() user is nil")
	}
	if registerResp.User.Email != "priya@test.com" {
		t.Fatalf("Register() email = %q, want normalized email", registerResp.User.Email)
	}
	if registerResp.Token == "" {
		t.Fatal("Register() token is empty")
	}

	loginResp, err := svc.Login(context.Background(), &rentrelaypb.LoginRequest{
		Email:    "priya@test.com",
		Password: "pass123",
	})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if loginResp.User.UserId != registerResp.User.UserId {
		t.Fatalf("Login() user_id = %q, want %q", loginResp.User.UserId, registerResp.User.UserId)
	}
}

func TestRegisterRejectsDuplicateEmail(t *testing.T) {
	svc := NewInMemoryService()
	req := &rentrelaypb.RegisterRequest{
		Name:     "Suresh Gupta",
		Email:    "suresh@test.com",
		Password: "pass123",
		Role:     rentrelaypb.UserRole_ROLE_LANDLORD,
	}

	if _, err := svc.Register(context.Background(), req); err != nil {
		t.Fatalf("Register() first call error = %v", err)
	}

	_, err := svc.Register(context.Background(), req)
	if status.Code(err) != codes.AlreadyExists {
		t.Fatalf("Register() duplicate code = %v, want %v", status.Code(err), codes.AlreadyExists)
	}
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	svc := NewInMemoryService()
	_, err := svc.Register(context.Background(), &rentrelaypb.RegisterRequest{
		Name:     "Priya Reddy",
		Email:    "priya@test.com",
		Password: "pass123",
		Role:     rentrelaypb.UserRole_ROLE_TENANT,
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	_, err = svc.Login(context.Background(), &rentrelaypb.LoginRequest{
		Email:    "priya@test.com",
		Password: "wrong",
	})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("Login() wrong password code = %v, want %v", status.Code(err), codes.Unauthenticated)
	}
}
