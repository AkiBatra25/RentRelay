package storageworker

import (
	"context"
	"hash/crc32"
	"io"
	"log"
	"sort"
	"strings"
	"sync"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const hashSlots = 256

type entry struct {
	value   []byte
	version int32
}

type Service struct {
	rentrelaypb.UnimplementedStorageWorkerServer

	workerID string
	mu       sync.RWMutex
	entries  map[string]entry
	wal      *WAL // the notebook — nil means no persistence (for tests)
}

// NewService creates a worker with NO disk persistence.
// Used in unit tests so tests don't create files on disk.
func NewService(workerID string) *Service {
	return &Service{workerID: workerID, entries: make(map[string]entry)}
}

// NewServiceWithWAL creates a worker WITH disk persistence.
// It opens the WAL file, replays all previous entries into memory,
// then continues appending new entries to the same file.
// This is what the real storage-worker binary uses.
func NewServiceWithWAL(workerID string, walPath string) (*Service, error) {
	wal, err := OpenWAL(walPath)
	if err != nil {
		return nil, err
	}

	// Replay the log — rebuild everything that was stored before the restart
	entries, err := wal.Replay()
	if err != nil {
		return nil, err
	}

	log.Printf("worker %s: replayed %d entries from WAL at %s", workerID, len(entries), walPath)

	return &Service{
		workerID: workerID,
		entries:  entries,
		wal:      wal,
	}, nil
}

func (s *Service) Put(ctx context.Context, req *rentrelaypb.KVPutRequest) (*rentrelaypb.KVPutResponse, error) {
	if req == nil || strings.TrimSpace(req.Key) == "" {
		return nil, status.Error(codes.InvalidArgument, "key is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	current, exists := s.entries[req.Key]
	if exists && req.ReplicaVersion < current.version {
		return nil, status.Error(codes.FailedPrecondition, "replica version is older than stored version")
	}

	// Write to disk BEFORE updating memory.
	// "Write-Ahead" means: log it first, then do it.
	// If we crash between the log write and memory write, on restart
	// the log has the entry and Replay() will restore it correctly.
	if s.wal != nil {
		if err := s.wal.LogPut(req.Key, req.Value, req.ReplicaVersion); err != nil {
			return nil, status.Errorf(codes.Internal, "wal write: %v", err)
		}
	}

	s.entries[req.Key] = entry{value: append([]byte(nil), req.Value...), version: req.ReplicaVersion}
	return &rentrelaypb.KVPutResponse{
		Success: true, WorkerId: s.workerID, ReplicaVersion: req.ReplicaVersion,
	}, nil
}

func (s *Service) Get(ctx context.Context, req *rentrelaypb.KVGetRequest) (*rentrelaypb.KVGetResponse, error) {
	if req == nil || strings.TrimSpace(req.Key) == "" {
		return nil, status.Error(codes.InvalidArgument, "key is required")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, found := s.entries[req.Key]
	return &rentrelaypb.KVGetResponse{
		Key: req.Key, Value: append([]byte(nil), value.value...), ReplicaVersion: value.version, Found: found,
	}, nil
}

func (s *Service) Delete(ctx context.Context, req *rentrelaypb.KVGetRequest) (*emptypb.Empty, error) {
	if req == nil || strings.TrimSpace(req.Key) == "" {
		return nil, status.Error(codes.InvalidArgument, "key is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	// Log the delete to disk before doing it in memory
	if s.wal != nil {
		if err := s.wal.LogDelete(req.Key); err != nil {
			return nil, status.Errorf(codes.Internal, "wal write: %v", err)
		}
	}

	delete(s.entries, req.Key)
	return &emptypb.Empty{}, nil
}

func (s *Service) ListKeys(ctx context.Context, req *rentrelaypb.KeyRangeRequest) (*rentrelaypb.KeyList, error) {
	if req == nil || req.ShardStart < 0 || req.ShardEnd >= hashSlots || req.ShardStart > req.ShardEnd {
		return nil, status.Error(codes.InvalidArgument, "invalid shard range")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	var keys []string
	for key := range s.entries {
		slot := int32(crc32.ChecksumIEEE([]byte(key)) % hashSlots)
		if slot >= req.ShardStart && slot <= req.ShardEnd {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return &rentrelaypb.KeyList{Keys: keys}, nil
}

func (s *Service) TransferKeys(stream grpc.ClientStreamingServer[rentrelaypb.KVPutRequest, emptypb.Empty]) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return stream.SendAndClose(&emptypb.Empty{})
			}
			return err
		}
		if _, err := s.Put(stream.Context(), req); err != nil {
			return err
		}
	}
}

func (s *Service) StoredKeys() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return int64(len(s.entries))
}

// Close shuts down the WAL cleanly when the service stops.
func (s *Service) Close() error {
	if s.wal != nil {
		return s.wal.Close()
	}
	return nil
}
