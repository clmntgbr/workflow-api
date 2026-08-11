package user

import (
	"context"
	"errors"

	domainuser "go-api/internal/domain/user"

	"github.com/google/uuid"
)

type GetUserByIDQuery struct {
	ID uuid.UUID
}

type GetUserByIDHandler struct {
	readRepo domainuser.UserReadRepository
}

func NewGetUserByIDHandler(readRepo domainuser.UserReadRepository) *GetUserByIDHandler {
	return &GetUserByIDHandler{readRepo: readRepo}
}

func (h *GetUserByIDHandler) Handle(ctx context.Context, q GetUserByIDQuery) (*domainuser.UserView, error) {
	view, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		return nil, errors.New("failed to get user")
	}
	if view == nil {
		return nil, errors.New("user not found")
	}
	return view, nil
}
