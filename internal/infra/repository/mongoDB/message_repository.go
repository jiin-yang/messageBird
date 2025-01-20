package mongoDB

import (
	"context"
	"fmt"
	"github.com/jiin-yang/messageBird/internal/message"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

const (
	messagesCollection = "messages"
	getMessageLimit    = 2
)

type repo struct {
	client     *Client
	collection *mongo.Collection
}

type NewMessageRepositoryOpts struct {
	Client *Client
}

func NewMessageRepository(opts *NewMessageRepositoryOpts) message.Repository {
	return &repo{
		client:     opts.Client,
		collection: opts.Client.Database.Collection(messagesCollection),
	}
}

func (r repo) CreateMessage(ctx context.Context, msgData message.CreateMessage) (*message.CreatedMessageDbResponse, error) {
	objID := bson.NewObjectID()
	timeNow := time.Now()

	dbData := Message{
		ID:          objID,
		PhoneNumber: msgData.PhoneNumber,
		Content:     msgData.Content,
		Status:      msgData.Status,
		CreatedAt:   &timeNow,
	}

	insertResult, err := r.collection.InsertOne(ctx, dbData)
	if err != nil {
		return nil, err
	}

	insertedID := insertResult.InsertedID.(bson.ObjectID)
	idStr := insertedID.Hex()

	createdMessage := message.CreatedMessageDbResponse{
		Id:          idStr,
		PhoneNumber: msgData.PhoneNumber,
		Content:     msgData.Content,
		Status:      msgData.Status,
		CreatedAt:   &timeNow,
	}

	return &createdMessage, nil
}

func (r repo) GetOldestStatusNewMessages(ctx context.Context) ([]message.Message, error) {
	filter := bson.M{"status": message.New}

	// NOT: Sort isleminde neden `_id` kullandim ?
	// mongoDB object id'si time bazli oldugu icin ve indexli oldugu icin createdAt yerine _id kullanmayi uygun gordum
	// Eger ki Create isleminin birden fazla server, client veya instance tarafindan yapilacagi bir senaryo olsa idi
	// createdAt alanini kullanirdim.
	findOpts := options.Find().
		SetSort(bson.D{{Key: "_id", Value: 1}}).
		SetLimit(getMessageLimit)

	cur, err := r.collection.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var result []message.Message

	for cur.Next(ctx) {
		var dbMsg Message
		if decodeErr := cur.Decode(&dbMsg); decodeErr != nil {
			return nil, decodeErr
		}

		msg := message.Message{
			Id:          dbMsg.ID.Hex(),
			PhoneNumber: dbMsg.PhoneNumber,
			Content:     dbMsg.Content,
			Status:      dbMsg.Status,
			CreatedAt:   dbMsg.CreatedAt,
		}

		result = append(result, msg)
	}

	if err = cur.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
func (r repo) UpdateMessageStatus(ctx context.Context, messageID string, newStatus message.Status) error {
	objID, err := bson.ObjectIDFromHex(messageID)
	if err != nil {
		return fmt.Errorf("invalid message ID: %w", err)
	}

	filter := bson.M{"_id": objID}
	update := bson.M{
		"$set": bson.M{
			"status":    newStatus,
			"updatedAt": time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update message status: %w", err)
	}

	return nil
}

func (r repo) GetSentStatusMessages(ctx context.Context) ([]message.Message, error) {
	filter := bson.M{"status": message.Sent}

	cur, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var result []message.Message

	for cur.Next(ctx) {
		var dbMsg Message
		if decodeErr := cur.Decode(&dbMsg); decodeErr != nil {
			return nil, decodeErr
		}

		msg := message.Message{
			Id:          dbMsg.ID.Hex(),
			PhoneNumber: dbMsg.PhoneNumber,
			Content:     dbMsg.Content,
			Status:      dbMsg.Status,
			CreatedAt:   dbMsg.CreatedAt,
		}

		result = append(result, msg)
	}

	if err = cur.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
