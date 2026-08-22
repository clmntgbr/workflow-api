package subscription

import (
	"context"
	"errors"

	domainsubscription "go-api/internal/domain/subscription"
	domainuser "go-api/internal/domain/user"

	"github.com/google/uuid"
)

var ErrSubscriptionNotFound = errors.New("subscription not found")

type GetCurrentSubscriptionQuery struct {
	UserID uuid.UUID
}

type GetCurrentSubscriptionHandler struct {
	userRepo         domainuser.UserReadRepository
	subscriptionRepo domainsubscription.SubscriptionReadRepository
}

func NewGetCurrentSubscriptionHandler(
	userRepo domainuser.UserReadRepository,
	subscriptionRepo domainsubscription.SubscriptionReadRepository,
) *GetCurrentSubscriptionHandler {
	return &GetCurrentSubscriptionHandler{
		userRepo:         userRepo,
		subscriptionRepo: subscriptionRepo,
	}
}

func (h *GetCurrentSubscriptionHandler) Handle(
	ctx context.Context,
	q GetCurrentSubscriptionQuery,
) (*domainsubscription.SubscriptionView, error) {
	user, err := h.userRepo.FindByID(ctx, q.UserID)
	if err != nil {
		return nil, errors.New("failed to get user")
	}
	if user == nil || user.SubscriptionID == nil {
		return nil, ErrSubscriptionNotFound
	}

	view, err := h.subscriptionRepo.FindByID(ctx, *user.SubscriptionID)
	if err != nil {
		return nil, errors.New("failed to get subscription")
	}
	if view == nil {
		return nil, ErrSubscriptionNotFound
	}

	return view, nil
}
