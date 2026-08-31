package handler

import (
	"context"

	cmdsubscription "go-api/internal/application/command/subscription"
	querysubscription "go-api/internal/application/query/subscription"
	domainsubscription "go-api/internal/domain/subscription"
)

type subscriptionGetCurrentHandler interface {
	Handle(ctx context.Context, q querysubscription.GetCurrentSubscriptionQuery) (*domainsubscription.SubscriptionView, error)
}

type subscriptionGetQuotaHandler interface {
	Handle(ctx context.Context, q querysubscription.GetQuotaUsageQuery) (*querysubscription.QuotaUsageView, error)
}

type subscriptionPreviewPlanChangeHandler interface {
	Handle(ctx context.Context, q querysubscription.PreviewPlanChangeQuery) (*querysubscription.PlanChangePreview, error)
}

type subscriptionCreateHandler interface {
	Handle(ctx context.Context, cmd cmdsubscription.CreateSubscriptionCommand) (*cmdsubscription.CreateSubscriptionResult, error)
}

type subscriptionCreateBillingPortalHandler interface {
	Handle(ctx context.Context, cmd cmdsubscription.CreateBillingPortalCommand) (string, error)
}
