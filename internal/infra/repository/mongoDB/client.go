package mongoDB

import (
	"github.com/jiin-yang/messageBird/config"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Client struct {
	Database *mongo.Database
}

func NewClient(config *config.MongoDBConfig) (c *Client, err error) {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(config.Host).SetServerAPIOptions(serverAPI)
	mongoClient, err := mongo.Connect(opts)
	if err != nil {
		return nil, err
	}

	db := mongoClient.Database(config.Name)
	client := Client{Database: db}

	return &client, nil
}
