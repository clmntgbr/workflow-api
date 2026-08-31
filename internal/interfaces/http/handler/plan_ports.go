package handler

import (
	"context"

	queryplan "go-api/internal/application/query/plan"
	domainplan "go-api/internal/domain/plan"
)

type planListActiveHandler interface {
	Handle(ctx context.Context, q queryplan.ListActivePlansQuery) ([]domainplan.PlanView, error)
}
