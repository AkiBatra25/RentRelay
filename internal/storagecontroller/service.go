package storagecontroller

import (
	"context"
	"hash/crc32"
	"sort"
	"strings"
	"sync"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const hashSlots = 256

type Service struct {
	rentrelaypb.UnimplementedStorageControllerServer

	mu         sync.RWMutex
	partitions map[string]*rentrelaypb.PartitionInfo
	version    int64
}

func NewService() *Service {
	return &Service{partitions: make(map[string]*rentrelaypb.PartitionInfo)}
}

func (s *Service) RegisterWorker(ctx context.Context, worker *rentrelaypb.PartitionInfo) (*emptypb.Empty, error) {
	if worker == nil || strings.TrimSpace(worker.WorkerId) == "" || strings.TrimSpace(worker.WorkerAddress) == "" {
		return nil, status.Error(codes.InvalidArgument, "worker_id and worker_address are required")
	}
	if worker.ShardStart < 0 || worker.ShardEnd >= hashSlots || worker.ShardStart > worker.ShardEnd {
		return nil, status.Error(codes.InvalidArgument, "invalid shard range")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	copy := clonePartition(worker)
	copy.IsAlive = true
	copy.LastHeartbeat = timestamppb.Now()
	s.partitions[copy.WorkerId] = copy
	s.version++
	return &emptypb.Empty{}, nil
}

func (s *Service) Heartbeat(ctx context.Context, req *rentrelaypb.HeartbeatRequest) (*emptypb.Empty, error) {
	if req == nil || strings.TrimSpace(req.WorkerId) == "" {
		return nil, status.Error(codes.InvalidArgument, "worker_id is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	worker, ok := s.partitions[req.WorkerId]
	if !ok {
		return nil, status.Error(codes.NotFound, "worker is not registered")
	}
	worker.IsAlive = true
	worker.LastHeartbeat = timestamppb.Now()
	return &emptypb.Empty{}, nil
}

func (s *Service) GetWorkerForKey(ctx context.Context, req *rentrelaypb.GetWorkerRequest) (*rentrelaypb.GetWorkerResponse, error) {
	if req == nil || strings.TrimSpace(req.Key) == "" {
		return nil, status.Error(codes.InvalidArgument, "key is required")
	}
	partitions := s.sortedAlivePartitions()
	if len(partitions) == 0 {
		return nil, status.Error(codes.Unavailable, "no storage workers are available")
	}

	slot := int32(crc32.ChecksumIEEE([]byte(req.Key)) % hashSlots)
	primaryIndex := 0
	for index, partition := range partitions {
		if slot >= partition.ShardStart && slot <= partition.ShardEnd {
			primaryIndex = index
			break
		}
	}

	replicaCount := 2
	if len(partitions)-1 < replicaCount {
		replicaCount = len(partitions) - 1
	}
	replicas := make([]*rentrelaypb.PartitionInfo, 0, replicaCount)
	for offset := 1; offset <= replicaCount; offset++ {
		replicas = append(replicas, clonePartition(partitions[(primaryIndex+offset)%len(partitions)]))
	}

	return &rentrelaypb.GetWorkerResponse{
		Primary:  clonePartition(partitions[primaryIndex]),
		Replicas: replicas,
	}, nil
}

func (s *Service) GetAllPartitions(context.Context, *emptypb.Empty) (*rentrelaypb.PartitionTable, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	partitions := make([]*rentrelaypb.PartitionInfo, 0, len(s.partitions))
	for _, partition := range s.partitions {
		partitions = append(partitions, clonePartition(partition))
	}
	sort.Slice(partitions, func(i, j int) bool { return partitions[i].ShardStart < partitions[j].ShardStart })
	return &rentrelaypb.PartitionTable{Partitions: partitions, Version: s.version}, nil
}

func (s *Service) WatchRebalance(req *rentrelaypb.HeartbeatRequest, stream grpc.ServerStreamingServer[rentrelaypb.RebalanceEvent]) error {
	if req == nil || strings.TrimSpace(req.WorkerId) == "" {
		return status.Error(codes.InvalidArgument, "worker_id is required")
	}
	<-stream.Context().Done()
	return stream.Context().Err()
}

func (s *Service) sortedAlivePartitions() []*rentrelaypb.PartitionInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	partitions := make([]*rentrelaypb.PartitionInfo, 0, len(s.partitions))
	for _, partition := range s.partitions {
		if partition.IsAlive {
			partitions = append(partitions, clonePartition(partition))
		}
	}
	sort.Slice(partitions, func(i, j int) bool { return partitions[i].ShardStart < partitions[j].ShardStart })
	return partitions
}

func clonePartition(partition *rentrelaypb.PartitionInfo) *rentrelaypb.PartitionInfo {
	if partition == nil {
		return nil
	}
	copy := *partition
	return &copy
}
