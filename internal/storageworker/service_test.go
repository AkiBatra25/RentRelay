package storageworker

import (
	"context"
	"os"
	"testing"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
)

// Original test — still works because NewService has no WAL
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

// New test — proves data survives a simulated restart using WAL
func TestWALSurvivesRestart(t *testing.T) {
	// Create a temporary file to act as our WAL (deleted after test)
	tmp, err := os.CreateTemp("", "wal-test-*.log")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	walPath := tmp.Name()
	tmp.Close()
	defer os.Remove(walPath) // clean up after test

	ctx := context.Background()

	// --- Simulate the FIRST run of the worker ---
	svc1, err := NewServiceWithWAL("worker-test", walPath)
	if err != nil {
		t.Fatalf("NewServiceWithWAL: %v", err)
	}

	// Save two keys
	if _, err := svc1.Put(ctx, &rentrelaypb.KVPutRequest{
		Key: "agreement-1", Value: []byte("signed-data"), ReplicaVersion: 1,
	}); err != nil {
		t.Fatalf("Put agreement-1: %v", err)
	}
	if _, err := svc1.Put(ctx, &rentrelaypb.KVPutRequest{
		Key: "agreement-2", Value: []byte("escrow-data"), ReplicaVersion: 1,
	}); err != nil {
		t.Fatalf("Put agreement-2: %v", err)
	}

	// Delete one key
	if _, err := svc1.Delete(ctx, &rentrelaypb.KVGetRequest{Key: "agreement-2"}); err != nil {
		t.Fatalf("Delete agreement-2: %v", err)
	}

	svc1.Close() // simulate worker shutting down

	// --- Simulate RESTART — new service, same WAL file ---
	svc2, err := NewServiceWithWAL("worker-test", walPath)
	if err != nil {
		t.Fatalf("NewServiceWithWAL after restart: %v", err)
	}
	defer svc2.Close()

	// agreement-1 should still be there
	got, err := svc2.Get(ctx, &rentrelaypb.KVGetRequest{Key: "agreement-1"})
	if err != nil {
		t.Fatalf("Get agreement-1: %v", err)
	}
	if !got.Found {
		t.Fatal("agreement-1 should exist after restart but was not found")
	}
	if string(got.Value) != "signed-data" {
		t.Fatalf("agreement-1 value = %q, want %q", got.Value, "signed-data")
	}

	// agreement-2 was deleted, should NOT be there
	got2, err := svc2.Get(ctx, &rentrelaypb.KVGetRequest{Key: "agreement-2"})
	if err != nil {
		t.Fatalf("Get agreement-2: %v", err)
	}
	if got2.Found {
		t.Fatal("agreement-2 was deleted but found after restart")
	}

	t.Log("WAL test passed — data survived simulated restart!")
}
