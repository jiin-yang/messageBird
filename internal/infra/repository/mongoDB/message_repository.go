package mongoDB

import (
	"context"
	"github.com/jiin-yang/messageBird/internal/message"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"time"
)

const (
	messagesCollection = "messages"
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

func (r repo) CreateMessage(ctx context.Context, msgData message.Message) (*message.CreatedMessage, error) {
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

	createdMessage := message.CreatedMessage{
		Id:          idStr,
		PhoneNumber: msgData.PhoneNumber,
		Content:     msgData.Content,
		Status:      msgData.Status,
		CreatedAt:   &timeNow,
	}

	return &createdMessage, nil
}
