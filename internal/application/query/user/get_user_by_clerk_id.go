package user

import (
	"context"
	"errors"

	domainuser "go-api/internal/domain/user"
)

type GetUserByClerkIDQuery struct {
	ClerkID string
}

type GetUserByClerkIDHandler struct {
	readRepo domainuser.UserReadRepository
}

func NewGetUserByClerkIDHandler(readRepo domainuser.UserReadRepository) *GetUserByClerkIDHandler {
	return &GetUserByClerkIDHandler{readRepo: readRepo}
}

func (h *GetUserByClerkIDHandler) Handle(ctx context.Context, q GetUserByClerkIDQuery) (*domainuser.UserView, error) {
	view, err := h.readRepo.FindByClerkID(ctx, q.ClerkID)
	if err != nil {
		return nil, errors.New("failed to get user")
	}
	return view, nil
}
