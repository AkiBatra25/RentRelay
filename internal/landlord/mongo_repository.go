package landlord

import (
	"context"
	"errors"
	"time"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoRepository struct {
	client     *mongo.Client
	collection *mongo.Collection
}

type leaseTermsDocument struct {
	LandlordID        string    `bson:"landlord_id"`
	PropertyID        string    `bson:"property_id"`
	LeaseDurationMo   int32     `bson:"lease_duration_mo"`
	NoticePeriodDays  int32     `bson:"notice_period_days"`
	PreferredTenant   string    `bson:"preferred_tenant"`
	AllowedTypes      []string  `bson:"allowed_types"`
	MaintenanceCharge float64   `bson:"maintenance_charge"`
	PaymentDueDay     string    `bson:"payment_due_day"`
	UpdatedAt         time.Time `bson:"updated_at"`
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
		collection: client.Database(database).Collection("lease_terms"),
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

func (r *MongoRepository) SaveLeaseTerms(ctx context.Context, terms *rentrelaypb.LeaseTerms) error {
	doc := toLeaseTermsDocument(terms)
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"landlord_id": terms.LandlordId, "property_id": terms.PropertyId},
		bson.M{"$set": doc},
		options.UpdateOne().SetUpsert(true),
	)
	return err
}

func (r *MongoRepository) FindLeaseTerms(ctx context.Context, landlordID string, propertyID string) (*rentrelaypb.LeaseTerms, error) {
	var doc leaseTermsDocument
	err := r.collection.FindOne(ctx, bson.M{"landlord_id": landlordID, "property_id": propertyID}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrLeaseTermsNotFound
	}
	if err != nil {
		return nil, err
	}
	return fromLeaseTermsDocument(doc), nil
}

func (r *MongoRepository) ensureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "property_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{Keys: bson.D{{Key: "landlord_id", Value: 1}}},
		{Keys: bson.D{{Key: "landlord_id", Value: 1}, {Key: "property_id", Value: 1}}},
	})
	return err
}

func toLeaseTermsDocument(terms *rentrelaypb.LeaseTerms) leaseTermsDocument {
	return leaseTermsDocument{
		LandlordID:        terms.LandlordId,
		PropertyID:        terms.PropertyId,
		LeaseDurationMo:   terms.LeaseDurationMo,
		NoticePeriodDays:  terms.NoticePeriodDays,
		PreferredTenant:   terms.PreferredTenant,
		AllowedTypes:      append([]string(nil), terms.AllowedTypes...),
		MaintenanceCharge: terms.MaintenanceCharge,
		PaymentDueDay:     terms.PaymentDueDay,
		UpdatedAt:         time.Now().UTC(),
	}
}

func fromLeaseTermsDocument(doc leaseTermsDocument) *rentrelaypb.LeaseTerms {
	return &rentrelaypb.LeaseTerms{
		LandlordId:        doc.LandlordID,
		PropertyId:        doc.PropertyID,
		LeaseDurationMo:   doc.LeaseDurationMo,
		NoticePeriodDays:  doc.NoticePeriodDays,
		PreferredTenant:   doc.PreferredTenant,
		AllowedTypes:      append([]string(nil), doc.AllowedTypes...),
		MaintenanceCharge: doc.MaintenanceCharge,
		PaymentDueDay:     doc.PaymentDueDay,
	}
}
