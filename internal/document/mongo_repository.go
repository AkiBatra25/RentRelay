package document

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

type documentRecord struct {
	DocumentID  string    `bson:"document_id"`
	AgreementID string    `bson:"agreement_id"`
	DocType     string    `bson:"doc_type"`
	StorageKey  string    `bson:"storage_key"`
	SHA256Hash  string    `bson:"sha256_hash"`
	SizeBytes   int64     `bson:"size_bytes"`
	UploadedBy  string    `bson:"uploaded_by"`
	Locked      bool      `bson:"locked"`
	Filename    string    `bson:"filename"`
	Content     []byte    `bson:"content"`
	UploadedAt  time.Time `bson:"uploaded_at"`
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
		collection: client.Database(database).Collection("documents"),
	}
	_, err = repo.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "document_id", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "agreement_id", Value: 1}}},
		{Keys: bson.D{{Key: "sha256_hash", Value: 1}}},
	})
	if err != nil {
		_ = client.Disconnect(ctx)
		return nil, err
	}
	return repo, nil
}

func (r *MongoRepository) Create(ctx context.Context, record *Record) error {
	_, err := r.collection.InsertOne(ctx, toDocumentRecord(record))
	return err
}

func (r *MongoRepository) FindByID(ctx context.Context, documentID string) (*Record, error) {
	return r.findOne(ctx, bson.M{"document_id": documentID})
}

func (r *MongoRepository) FindFirstByAgreement(ctx context.Context, agreementID string) (*Record, error) {
	return r.findOne(ctx, bson.M{"agreement_id": agreementID})
}

func (r *MongoRepository) ListByAgreement(ctx context.Context, agreementID string) ([]*rentrelaypb.Document, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"agreement_id": agreementID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var documents []*rentrelaypb.Document
	for cursor.Next(ctx) {
		var record documentRecord
		if err := cursor.Decode(&record); err != nil {
			return nil, err
		}
		documents = append(documents, fromDocumentRecord(record).Document)
	}
	return documents, cursor.Err()
}

func (r *MongoRepository) SetLockedByAgreement(ctx context.Context, agreementID string, locked bool) error {
	_, err := r.collection.UpdateMany(ctx, bson.M{"agreement_id": agreementID}, bson.M{"$set": bson.M{"locked": locked}})
	return err
}

func (r *MongoRepository) Close(ctx context.Context) error {
	return r.client.Disconnect(ctx)
}

func (r *MongoRepository) findOne(ctx context.Context, filter bson.M) (*Record, error) {
	var record documentRecord
	err := r.collection.FindOne(ctx, filter).Decode(&record)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrDocumentNotFound
	}
	if err != nil {
		return nil, err
	}
	return fromDocumentRecord(record), nil
}

func toDocumentRecord(record *Record) documentRecord {
	return documentRecord{
		DocumentID:  record.Document.DocumentId,
		AgreementID: record.Document.AgreementId,
		DocType:     record.Document.DocType,
		StorageKey:  record.Document.StorageKey,
		SHA256Hash:  record.Document.Sha256Hash,
		SizeBytes:   record.Document.SizeBytes,
		UploadedBy:  record.Document.UploadedBy,
		Locked:      record.Document.Locked,
		Filename:    record.Filename,
		Content:     append([]byte(nil), record.Content...),
		UploadedAt:  record.Document.UploadedAt.AsTime(),
	}
}

func fromDocumentRecord(record documentRecord) *Record {
	return &Record{
		Document: &rentrelaypb.Document{
			DocumentId:  record.DocumentID,
			AgreementId: record.AgreementID,
			DocType:     record.DocType,
			StorageKey:  record.StorageKey,
			Sha256Hash:  record.SHA256Hash,
			SizeBytes:   record.SizeBytes,
			UploadedBy:  record.UploadedBy,
			Locked:      record.Locked,
			UploadedAt:  timestamppb.New(record.UploadedAt),
		},
		Filename: record.Filename,
		Content:  append([]byte(nil), record.Content...),
	}
}
