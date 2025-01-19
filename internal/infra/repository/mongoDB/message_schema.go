package mongoDB

import "go.mongodb.org/mongo-driver/v2/bson"

type Message struct {
	ID          *bson.ObjectID `bson:"_id"`
	PhoneNumber string         `bson:"phone_number"`
	Content     string         `bson:"content"`
	Status      string         `bson:"status"`
}
