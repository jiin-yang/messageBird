package message

import (
	"context"
)

type Repository interface {
	CreateMessage(ctx context.Context, message CreateMessage) (*CreatedMessageDbResponse, error)
	GetOldestStatusNewMessages(ctx context.Context) ([]Message, error)
}
