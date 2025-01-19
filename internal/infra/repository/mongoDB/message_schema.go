package mongoDB

import (
	"github.com/jiin-yang/messageBird/internal/message"
	"go.mongodb.org/mongo-driver/v2/bson"
	"time"
)

type Message struct {
	ID          bson.ObjectID  `bson:"_id"`
	PhoneNumber string         `bson:"phoneNumber"`
	Content     string         `bson:"content"`
	Status      message.Status `bson:"status"`
	CreatedAt   *time.Time     `bson:"createdAt"`
	UpdatedAt   *time.Time     `bson:"updatedAt,omitempty"`
}
