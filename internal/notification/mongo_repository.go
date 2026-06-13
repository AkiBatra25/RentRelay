package notification

import (
	"context"
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

type notificationDocument struct {
	NotificationID string    `bson:"notification_id"`
	UserID         string    `bson:"user_id"`
	AgreementID    string    `bson:"agreement_id"`
	Event          int32     `bson:"event"`
	Channel        int32     `bson:"channel"`
	Message        string    `bson:"message"`
	Delivered      bool      `bson:"delivered"`
	CreatedAt      time.Time `bson:"created_at"`
	DeliveredAt    time.Time `bson:"delivered_at"`
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
	repo := &MongoRepository{client: client, collection: client.Database(database).Collection("notifications")}
	_, err = repo.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "notification_id", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "created_at", Value: -1}}},
		{Keys: bson.D{{Key: "created_at", Value: 1}}, Options: options.Index().SetExpireAfterSeconds(7776000)},
	})
	if err != nil {
		_ = client.Disconnect(ctx)
		return nil, err
	}
	return repo, nil
}

func (r *MongoRepository) Create(ctx context.Context, notification *rentrelaypb.Notification) error {
	_, err := r.collection.InsertOne(ctx, toNotificationDocument(notification))
	return err
}

func (r *MongoRepository) ListByUser(ctx context.Context, userID string) ([]*rentrelaypb.Notification, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var results []*rentrelaypb.Notification
	for cursor.Next(ctx) {
		var item notificationDocument
		if err := cursor.Decode(&item); err != nil {
			return nil, err
		}
		results = append(results, fromNotificationDocument(item))
	}
	return results, cursor.Err()
}

func (r *MongoRepository) Close(ctx context.Context) error { return r.client.Disconnect(ctx) }

func toNotificationDocument(value *rentrelaypb.Notification) notificationDocument {
	return notificationDocument{
		NotificationID: value.NotificationId, UserID: value.UserId, AgreementID: value.AgreementId,
		Event: int32(value.Event), Channel: int32(value.Channel), Message: value.Message,
		Delivered: value.Delivered, CreatedAt: value.CreatedAt.AsTime(), DeliveredAt: value.DeliveredAt.AsTime(),
	}
}

func fromNotificationDocument(value notificationDocument) *rentrelaypb.Notification {
	return &rentrelaypb.Notification{
		NotificationId: value.NotificationID, UserId: value.UserID, AgreementId: value.AgreementID,
		Event: rentrelaypb.NotificationEvent(value.Event), Channel: rentrelaypb.NotificationChannel(value.Channel),
		Message: value.Message, Delivered: value.Delivered,
		CreatedAt: timestamppb.New(value.CreatedAt), DeliveredAt: timestamppb.New(value.DeliveredAt),
	}
}
