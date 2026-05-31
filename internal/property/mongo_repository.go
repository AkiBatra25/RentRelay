package property

import (
	"context"
	"errors"
	"regexp"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoRepository struct {
	client     *mongo.Client
	collection *mongo.Collection
}

type propertyDocument struct {
	PropertyID  string   `bson:"property_id"`
	LandlordID  string   `bson:"landlord_id"`
	Title       string   `bson:"title"`
	Address     string   `bson:"address"`
	City        string   `bson:"city"`
	Zone        string   `bson:"zone"`
	Latitude    float64  `bson:"latitude"`
	Longitude   float64  `bson:"longitude"`
	Bedrooms    int32    `bson:"bedrooms"`
	RentMonthly float64  `bson:"rent_monthly"`
	DepositAmt  float64  `bson:"deposit_amt"`
	Furnishing  int32    `bson:"furnishing"`
	Amenities   []string `bson:"amenities"`
	IsAvailable bool     `bson:"is_available"`
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
		collection: client.Database(database).Collection("properties"),
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

func (r *MongoRepository) Create(ctx context.Context, property *rentrelaypb.Property) error {
	_, err := r.collection.InsertOne(ctx, toDocument(property))
	if err == nil {
		return nil
	}
	if mongo.IsDuplicateKeyError(err) {
		return ErrDuplicateProperty
	}
	return err
}

func (r *MongoRepository) FindByID(ctx context.Context, propertyID string) (*rentrelaypb.Property, error) {
	return r.findOne(ctx, bson.M{"property_id": propertyID})
}

func (r *MongoRepository) Search(ctx context.Context, filter SearchFilter) ([]*rentrelaypb.Property, error) {
	query := bson.M{"is_available": true}

	if filter.City != "" {
		query["city"] = caseInsensitiveExact(filter.City)
	}
	if filter.Zone != "" {
		query["zone"] = caseInsensitiveExact(filter.Zone)
	}
	if filter.MinBedrooms > 0 {
		query["bedrooms"] = bson.M{"$gte": filter.MinBedrooms}
	}
	if filter.MaxRent > 0 {
		query["rent_monthly"] = bson.M{"$lte": filter.MaxRent}
	}
	if filter.Furnishing != rentrelaypb.FurnishingType_FURNISHING_UNKNOWN {
		query["furnishing"] = int32(filter.Furnishing)
	}

	cursor, err := r.collection.Find(ctx, query)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*rentrelaypb.Property
	for cursor.Next(ctx) {
		var doc propertyDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		results = append(results, fromDocument(doc))
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *MongoRepository) UpdateAvailability(ctx context.Context, propertyID string, isAvailable bool) (*rentrelaypb.Property, error) {
	var doc propertyDocument
	err := r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"property_id": propertyID},
		bson.M{"$set": bson.M{"is_available": isAvailable}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&doc)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrPropertyNotFound
	}
	if err != nil {
		return nil, err
	}

	return fromDocument(doc), nil
}

func (r *MongoRepository) ListByLandlord(ctx context.Context, landlordID string) ([]*rentrelaypb.Property, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"landlord_id": landlordID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*rentrelaypb.Property
	for cursor.Next(ctx) {
		var doc propertyDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		results = append(results, fromDocument(doc))
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *MongoRepository) ensureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "property_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{Keys: bson.D{{Key: "landlord_id", Value: 1}}},
		{Keys: bson.D{{Key: "city", Value: 1}, {Key: "zone", Value: 1}}},
		{Keys: bson.D{{Key: "rent_monthly", Value: 1}}},
		{Keys: bson.D{{Key: "is_available", Value: 1}}},
	})
	return err
}

func (r *MongoRepository) findOne(ctx context.Context, filter bson.M) (*rentrelaypb.Property, error) {
	var doc propertyDocument
	err := r.collection.FindOne(ctx, filter).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrPropertyNotFound
	}
	if err != nil {
		return nil, err
	}
	return fromDocument(doc), nil
}

func toDocument(property *rentrelaypb.Property) propertyDocument {
	return propertyDocument{
		PropertyID:  property.PropertyId,
		LandlordID:  property.LandlordId,
		Title:       property.Title,
		Address:     property.Address,
		City:        property.City,
		Zone:        property.Zone,
		Latitude:    property.Latitude,
		Longitude:   property.Longitude,
		Bedrooms:    property.Bedrooms,
		RentMonthly: property.RentMonthly,
		DepositAmt:  property.DepositAmt,
		Furnishing:  int32(property.Furnishing),
		Amenities:   append([]string(nil), property.Amenities...),
		IsAvailable: property.IsAvailable,
	}
}

func fromDocument(doc propertyDocument) *rentrelaypb.Property {
	return &rentrelaypb.Property{
		PropertyId:  doc.PropertyID,
		LandlordId:  doc.LandlordID,
		Title:       doc.Title,
		Address:     doc.Address,
		City:        doc.City,
		Zone:        doc.Zone,
		Latitude:    doc.Latitude,
		Longitude:   doc.Longitude,
		Bedrooms:    doc.Bedrooms,
		RentMonthly: doc.RentMonthly,
		DepositAmt:  doc.DepositAmt,
		Furnishing:  rentrelaypb.FurnishingType(doc.Furnishing),
		Amenities:   append([]string(nil), doc.Amenities...),
		IsAvailable: doc.IsAvailable,
	}
}

func caseInsensitiveExact(value string) bson.M {
	return bson.M{
		"$regex":   "^" + regexp.QuoteMeta(value) + "$",
		"$options": "i",
	}
}
