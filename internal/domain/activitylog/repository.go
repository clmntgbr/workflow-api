package activitylog

import (
	"context"

	"go-api/internal/domain/paginate"

	"github.com/google/uuid"
)

type WriteRepository interface {
	Save(ctx context.Context, entry *Entry) error
}

type ReadRepository interface {
	FindByWorkflowID(
		ctx context.Context,
		workflowID uuid.UUID,
		query paginate.PaginateQuery,
	) ([]View, int64, error)
}
