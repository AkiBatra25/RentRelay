package storagecontroller

import (
	"context"
	"testing"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
)

func TestRoutesKeyToPrimaryAndTwoReplicas(t *testing.T) {
	svc := NewService()
	for index := int32(0); index < 4; index++ {
		_, err := svc.RegisterWorker(context.Background(), &rentrelaypb.PartitionInfo{
			WorkerId: "worker-" + string(rune('0'+index)), WorkerAddress: "worker",
			ShardStart: index * 64, ShardEnd: index*64 + 63,
		})
		if err != nil {
			t.Fatalf("RegisterWorker() error = %v", err)
		}
	}

	route, err := svc.GetWorkerForKey(context.Background(), &rentrelaypb.GetWorkerRequest{Key: "agreement-123"})
	if err != nil {
		t.Fatalf("GetWorkerForKey() error = %v", err)
	}
	if route.Primary == nil || len(route.Replicas) != 2 {
		t.Fatalf("route = %#v, want primary and two replicas", route)
	}
}
