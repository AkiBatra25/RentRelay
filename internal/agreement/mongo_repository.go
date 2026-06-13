package agreement

import (
	"context"
	"errors"
	"time"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MongoRepository struct {
	client     *mongo.Client
	collection *mongo.Collection
}

type agreementDocument struct {
	AgreementID    string    `bson:"agreement_id"`
	TenantID       string    `bson:"tenant_id"`
	LandlordID     string    `bson:"landlord_id"`
	PropertyID     string    `bson:"property_id"`
	State          int32     `bson:"state"`
	MonthlyRent    float64   `bson:"monthly_rent"`
	DepositAmount  float64   `bson:"deposit_amount"`
	DepositHeld    float64   `bson:"deposit_held"`
	LeaseMonths    int32     `bson:"lease_months"`
	NoticeDays     int32     `bson:"notice_days"`
	DocumentHash   string    `bson:"document_hash"`
	Signatures     []string  `bson:"signatures"`
	StartDate      time.Time `bson:"start_date"`
	EndDate        time.Time `bson:"end_date"`
	CreatedAt      time.Time `bson:"created_at"`
	UpdatedAt      time.Time `bson:"updated_at"`
	WorkerNode     string    `bson:"worker_node"`
	ReplicaVersion int32     `bson:"replica_version"`
}

func NewMongoRepository(ctx context.Context, uri string, database string) (*MongoRepository, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, err
	}

	repo := &MongoRepository{
		client:     client,
		collection: client.Database(database).Collection("agreements"),
	}
	if err := repo.ensureIndexes(ctx); err != nil {
		_ = client.Disconnect(ctx)
		return nil, err
	}
	return repo, nil
}

func (r *MongoRepository) Close(ctx context.Context) error {
	return r.client.Disconnect(ctx)
}

func (r *MongoRepository) Create(ctx context.Context, agreement *rentrelaypb.Agreement) error {
	_, err := r.collection.InsertOne(ctx, toAgreementDocument(agreement))
	return err
}

func (r *MongoRepository) FindByID(ctx context.Context, agreementID string) (*rentrelaypb.Agreement, error) {
	var doc agreementDocument
	err := r.collection.FindOne(ctx, bson.M{"agreement_id": agreementID}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrAgreementNotFound
	}
	if err != nil {
		return nil, err
	}
	return fromAgreementDocument(doc), nil
}

func (r *MongoRepository) Save(ctx context.Context, agreement *rentrelaypb.Agreement) error {
	result, err := r.collection.ReplaceOne(
		ctx,
		bson.M{"agreement_id": agreement.AgreementId},
		toAgreementDocument(agreement),
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrAgreementNotFound
	}
	return nil
}

func (r *MongoRepository) ensureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "agreement_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{Keys: bson.D{{Key: "tenant_id", Value: 1}}},
		{Keys: bson.D{{Key: "landlord_id", Value: 1}}},
		{Keys: bson.D{{Key: "property_id", Value: 1}}},
		{Keys: bson.D{{Key: "state", Value: 1}}},
	})
	return err
}

func toAgreementDocument(agreement *rentrelaypb.Agreement) agreementDocument {
	return agreementDocument{
		AgreementID:    agreement.AgreementId,
		TenantID:       agreement.TenantId,
		LandlordID:     agreement.LandlordId,
		PropertyID:     agreement.PropertyId,
		State:          int32(agreement.State),
		MonthlyRent:    agreement.MonthlyRent,
		DepositAmount:  agreement.DepositAmount,
		DepositHeld:    agreement.DepositHeld,
		LeaseMonths:    agreement.LeaseMonths,
		NoticeDays:     agreement.NoticeDays,
		DocumentHash:   agreement.DocumentHash,
		Signatures:     append([]string(nil), agreement.Signatures...),
		StartDate:      timestampTime(agreement.StartDate),
		EndDate:        timestampTime(agreement.EndDate),
		CreatedAt:      timestampTime(agreement.CreatedAt),
		UpdatedAt:      timestampTime(agreement.UpdatedAt),
		WorkerNode:     agreement.WorkerNode,
		ReplicaVersion: agreement.ReplicaVersion,
	}
}

func fromAgreementDocument(doc agreementDocument) *rentrelaypb.Agreement {
	return &rentrelaypb.Agreement{
		AgreementId:    doc.AgreementID,
		TenantId:       doc.TenantID,
		LandlordId:     doc.LandlordID,
		PropertyId:     doc.PropertyID,
		State:          rentrelaypb.AgreementState(doc.State),
		MonthlyRent:    doc.MonthlyRent,
		DepositAmount:  doc.DepositAmount,
		DepositHeld:    doc.DepositHeld,
		LeaseMonths:    doc.LeaseMonths,
		NoticeDays:     doc.NoticeDays,
		DocumentHash:   doc.DocumentHash,
		Signatures:     append([]string(nil), doc.Signatures...),
		StartDate:      timeTimestamp(doc.StartDate),
		EndDate:        timeTimestamp(doc.EndDate),
		CreatedAt:      timeTimestamp(doc.CreatedAt),
		UpdatedAt:      timeTimestamp(doc.UpdatedAt),
		WorkerNode:     doc.WorkerNode,
		ReplicaVersion: doc.ReplicaVersion,
	}
}

func timestampTime(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return ts.AsTime()
}

func timeTimestamp(value time.Time) *timestamppb.Timestamp {
	if value.IsZero() {
		return nil
	}
	return timestamppb.New(value)
}
