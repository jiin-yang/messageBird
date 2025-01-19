package message

import (
	"context"
)

type Repository interface {
	CreateMessage(ctx context.Context, message Message) (*CreatedMessage, error)
}
