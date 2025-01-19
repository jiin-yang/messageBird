package mongoDB

import (
	"github.com/jiin-yang/messageBird/internal/message"
	"go.mongodb.org/mongo-driver/v2/mongo"
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
