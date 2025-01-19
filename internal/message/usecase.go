package message

import "context"

type UseCase interface {
	CreateMessage(ctx context.Context, requestMsg CreateMessageRequest) (*CreateMessageResponse, error)
}

type useCase struct {
	repo Repository
}

type NewUseCaseOptions struct {
	Repo Repository
}

func NewUseCase(opts *NewUseCaseOptions) UseCase {
	return &useCase{
		repo: opts.Repo,
	}
}

func (u useCase) CreateMessage(ctx context.Context, requestMsg CreateMessageRequest) (*CreateMessageResponse, error) {
	msg := Message{
		PhoneNumber: requestMsg.PhoneNumber,
		Content:     requestMsg.Content,
		Status:      New,
	}

	dbRes, err := u.repo.CreateMessage(ctx, msg)
	if err != nil {
		return nil, err
	}

	createdMsgRes := CreateMessageResponse{
		Id:          dbRes.Id,
		PhoneNumber: dbRes.PhoneNumber,
		Content:     dbRes.Content,
		Status:      dbRes.Status.String(),
		CreatedAt:   dbRes.CreatedAt,
	}

	return &createdMsgRes, err
}
