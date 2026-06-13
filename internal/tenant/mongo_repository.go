package tenant

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

type rentalRequestDocument struct {
	RequestID      string    `bson:"request_id"`
	TenantID       string    `bson:"tenant_id"`
	PreferredZone  string    `bson:"preferred_zone"`
	PreferredCity  string    `bson:"preferred_city"`
	BedroomsNeeded int32     `bson:"bedrooms_needed"`
	MaxRent        float64   `bson:"max_rent"`
	Furnishing     int32     `bson:"furnishing"`
	MoveInDate     time.Time `bson:"move_in_date"`
	CreatedAt      time.Time `bson:"created_at"`
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
		collection: client.Database(database).Collection("rental_requests"),
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

func (r *MongoRepository) CreateRentalRequest(ctx context.Context, request *rentrelaypb.RentalRequest) error {
	doc := toRentalRequestDocument(request)
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"tenant_id": request.TenantId},
		bson.M{"$set": doc},
		options.UpdateOne().SetUpsert(true),
	)
	return err
}

func (r *MongoRepository) FindRentalRequestByTenant(ctx context.Context, tenantID string) (*rentrelaypb.RentalRequest, error) {
	var doc rentalRequestDocument
	err := r.collection.FindOne(ctx, bson.M{"tenant_id": tenantID}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrRentalRequestNotFound
	}
	if err != nil {
		return nil, err
	}
	return fromRentalRequestDocument(doc), nil
}

func (r *MongoRepository) ensureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "request_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "tenant_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{Keys: bson.D{{Key: "preferred_city", Value: 1}, {Key: "preferred_zone", Value: 1}}},
		{Keys: bson.D{{Key: "max_rent", Value: 1}}},
	})
	return err
}

func toRentalRequestDocument(request *rentrelaypb.RentalRequest) rentalRequestDocument {
	return rentalRequestDocument{
		RequestID:      request.RequestId,
		TenantID:       request.TenantId,
		PreferredZone:  request.PreferredZone,
		PreferredCity:  request.PreferredCity,
		BedroomsNeeded: request.BedroomsNeeded,
		MaxRent:        request.MaxRent,
		Furnishing:     int32(request.Furnishing),
		MoveInDate:     request.MoveInDate.AsTime(),
		CreatedAt:      request.CreatedAt.AsTime(),
	}
}

func fromRentalRequestDocument(doc rentalRequestDocument) *rentrelaypb.RentalRequest {
	return &rentrelaypb.RentalRequest{
		RequestId:      doc.RequestID,
		TenantId:       doc.TenantID,
		PreferredZone:  doc.PreferredZone,
		PreferredCity:  doc.PreferredCity,
		BedroomsNeeded: doc.BedroomsNeeded,
		MaxRent:        doc.MaxRent,
		Furnishing:     rentrelaypb.FurnishingType(doc.Furnishing),
		MoveInDate:     timestamppb.New(doc.MoveInDate),
		CreatedAt:      timestamppb.New(doc.CreatedAt),
	}
}
