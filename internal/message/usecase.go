package message

type UseCase interface {
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
