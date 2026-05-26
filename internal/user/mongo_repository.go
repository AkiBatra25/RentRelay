package user

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

type userDocument struct {
	UserID       string    `bson:"user_id"`
	Name         string    `bson:"name"`
	Email        string    `bson:"email"`
	Phone        string    `bson:"phone"`
	Role         int32     `bson:"role"`
	AadhaarHash  string    `bson:"aadhaar_hash"`
	KYCVerified  bool      `bson:"kyc_verified"`
	PasswordHash string    `bson:"password_hash"`
	CreatedAt    time.Time `bson:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at"`
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
		collection: client.Database(database).Collection("users"),
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

func (r *MongoRepository) Create(ctx context.Context, record *Record) error {
	_, err := r.collection.InsertOne(ctx, toDocument(record))
	if err == nil {
		return nil
	}
	if mongo.IsDuplicateKeyError(err) {
		return ErrDuplicateEmail
	}
	return err
}

func (r *MongoRepository) FindByEmail(ctx context.Context, email string) (*Record, error) {
	return r.findOne(ctx, bson.M{"email": email})
}

func (r *MongoRepository) FindByID(ctx context.Context, userID string) (*Record, error) {
	return r.findOne(ctx, bson.M{"user_id": userID})
}

func (r *MongoRepository) UpdateKYC(ctx context.Context, userID string, aadhaarHash string, verified bool) (*Record, error) {
	update := bson.M{
		"$set": bson.M{
			"aadhaar_hash": aadhaarHash,
			"kyc_verified": verified,
			"updated_at":   time.Now().UTC(),
		},
	}

	var doc userDocument
	err := r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"user_id": userID},
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return fromDocument(doc), nil
}

func (r *MongoRepository) ensureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	})
	return err
}

func (r *MongoRepository) findOne(ctx context.Context, filter bson.M) (*Record, error) {
	var doc userDocument
	err := r.collection.FindOne(ctx, filter).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return fromDocument(doc), nil
}

func toDocument(record *Record) userDocument {
	return userDocument{
		UserID:       record.User.UserId,
		Name:         record.User.Name,
		Email:        record.User.Email,
		Phone:        record.User.Phone,
		Role:         int32(record.User.Role),
		AadhaarHash:  record.User.AadhaarHash,
		KYCVerified:  record.User.KycVerified,
		PasswordHash: record.PasswordHash,
		CreatedAt:    record.User.CreatedAt.AsTime(),
		UpdatedAt:    record.User.UpdatedAt.AsTime(),
	}
}

func fromDocument(doc userDocument) *Record {
	return &Record{
		User: &rentrelaypb.User{
			UserId:      doc.UserID,
			Name:        doc.Name,
			Email:       doc.Email,
			Phone:       doc.Phone,
			Role:        rentrelaypb.UserRole(doc.Role),
			AadhaarHash: doc.AadhaarHash,
			KycVerified: doc.KYCVerified,
			CreatedAt:   timestamppb.New(doc.CreatedAt),
			UpdatedAt:   timestamppb.New(doc.UpdatedAt),
		},
		PasswordHash: doc.PasswordHash,
	}
}
