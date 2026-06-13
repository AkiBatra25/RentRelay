package agreement

import (
	"context"
	"errors"
	"fmt"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
)

const writeQuorum = 2

type StorageReplicator struct {
	controller rentrelaypb.StorageControllerClient
}

func NewStorageReplicator(controller rentrelaypb.StorageControllerClient) *StorageReplicator {
	return &StorageReplicator{controller: controller}
}

func (r *StorageReplicator) Store(ctx context.Context, agreement *rentrelaypb.Agreement) error {
	route, err := r.controller.GetWorkerForKey(ctx, &rentrelaypb.GetWorkerRequest{Key: agreement.AgreementId})
	if err != nil {
		return fmt.Errorf("route agreement: %w", err)
	}
	if route.Primary == nil {
		return errors.New("storage controller returned no primary worker")
	}

	agreement.WorkerNode = route.Primary.WorkerId
	value, err := proto.Marshal(agreement)
	if err != nil {
		return fmt.Errorf("serialize agreement: %w", err)
	}

	targets := append([]*rentrelaypb.PartitionInfo{route.Primary}, route.Replicas...)
	acknowledgements := 0
	for index, target := range targets {
		if target == nil || target.WorkerAddress == "" {
			continue
		}
		if err := putAgreement(ctx, target.WorkerAddress, agreement, value, index == 0); err == nil {
			acknowledgements++
		}
	}
	if acknowledgements < writeQuorum {
		return fmt.Errorf("write quorum not reached: got %d acknowledgements, need %d", acknowledgements, writeQuorum)
	}
	return nil
}

func (r *StorageReplicator) Load(ctx context.Context, agreementID string) (*rentrelaypb.Agreement, error) {
	route, err := r.controller.GetWorkerForKey(ctx, &rentrelaypb.GetWorkerRequest{Key: agreementID})
	if err != nil {
		return nil, fmt.Errorf("route agreement: %w", err)
	}

	targets := append([]*rentrelaypb.PartitionInfo{route.Primary}, route.Replicas...)
	for _, target := range targets {
		if target == nil || target.WorkerAddress == "" {
			continue
		}
		agreement, found, err := getAgreement(ctx, target.WorkerAddress, agreementID)
		if err == nil && found {
			return agreement, nil
		}
	}
	return nil, ErrAgreementNotFound
}

func putAgreement(ctx context.Context, address string, agreement *rentrelaypb.Agreement, value []byte, primary bool) error {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	response, err := rentrelaypb.NewStorageWorkerClient(conn).Put(ctx, &rentrelaypb.KVPutRequest{
		Key:            agreement.AgreementId,
		Value:          value,
		ReplicaVersion: agreement.ReplicaVersion,
		IsPrimary:      primary,
	})
	if err != nil {
		return err
	}
	if !response.Success {
		return errors.New("storage worker rejected agreement")
	}
	return nil
}

func getAgreement(ctx context.Context, address string, agreementID string) (*rentrelaypb.Agreement, bool, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, false, err
	}
	defer conn.Close()

	response, err := rentrelaypb.NewStorageWorkerClient(conn).Get(ctx, &rentrelaypb.KVGetRequest{Key: agreementID})
	if err != nil {
		return nil, false, err
	}
	if !response.Found {
		return nil, false, nil
	}

	var agreement rentrelaypb.Agreement
	if err := proto.Unmarshal(response.Value, &agreement); err != nil {
		return nil, false, fmt.Errorf("deserialize agreement: %w", err)
	}
	return &agreement, true, nil
}
