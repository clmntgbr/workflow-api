package identity

import (
	"context"

	"go-api/internal/domain/port"
)

type FetchUserHandler struct {
	userFetcher port.ClerkUserFetcher
}

func NewFetchUserHandler(userFetcher port.ClerkUserFetcher) *FetchUserHandler {
	return &FetchUserHandler{userFetcher: userFetcher}
}

func (s *FetchUserHandler) Handle(ctx context.Context, externalID string) (port.ClerkUser, error) {
	return s.userFetcher.Get(ctx, externalID)
}
