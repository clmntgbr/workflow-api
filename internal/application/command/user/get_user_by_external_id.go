package user

import (
	"context"
	"errors"

	domainuser "go-api/internal/domain/user"
)

// GetUserByExternalIDHandler loads the write-side aggregate (used by webhook update path).
type GetUserByExternalIDHandler struct {
	repo domainuser.UserWriteRepository
}

func NewGetUserByExternalIDHandler(repo domainuser.UserWriteRepository) *GetUserByExternalIDHandler {
	return &GetUserByExternalIDHandler{repo: repo}
}

func (h *GetUserByExternalIDHandler) Handle(ctx context.Context, externalID string) (*domainuser.User, error) {
	u, err := h.repo.GetByClerkID(ctx, externalID)
	if err != nil {
		return nil, errors.New("failed to get user by external ID")
	}
	return u, nil
}
