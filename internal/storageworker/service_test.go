package storageworker

import (
	"context"
	"testing"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
)

func TestPutGetAndVersionCheck(t *testing.T) {
	svc := NewService("worker-0")
	ctx := context.Background()

	if _, err := svc.Put(ctx, &rentrelaypb.KVPutRequest{Key: "key-1", Value: []byte("value"), ReplicaVersion: 2}); err != nil {
		t.Fatalf("Put() error = %v", err)
	}
	got, err := svc.Get(ctx, &rentrelaypb.KVGetRequest{Key: "key-1"})
	if err != nil || !got.Found || string(got.Value) != "value" {
		t.Fatalf("Get() = %#v, error = %v", got, err)
	}
	if _, err := svc.Put(ctx, &rentrelaypb.KVPutRequest{Key: "key-1", Value: []byte("old"), ReplicaVersion: 1}); err == nil {
		t.Fatal("older replica version was accepted")
	}
}
