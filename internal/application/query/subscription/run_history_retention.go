package subscription

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type RunHistoryCutoffResolver interface {
	RunHistoryCutoff(ctx context.Context, userID uuid.UUID, projectID uuid.UUID) (*time.Time, error)
}

func RunHistoryCutoff(retentionDays int, now time.Time) *time.Time {
	if retentionDays <= 0 {
		return nil
	}
	cutoff := now.UTC().AddDate(0, 0, -retentionDays)
	return &cutoff
}

func ClampTimeFrom(requested, cutoff *time.Time) *time.Time {
	if cutoff == nil {
		return requested
	}
	if requested == nil || requested.Before(*cutoff) {
		return cutoff
	}
	return requested
}

func IsWithinRunHistory(createdAt time.Time, cutoff *time.Time) bool {
	if cutoff == nil {
		return true
	}
	return !createdAt.Before(*cutoff)
}

func (h *GetQuotaUsageHandler) RunHistoryCutoff(
	ctx context.Context,
	userID uuid.UUID,
	projectID uuid.UUID,
) (*time.Time, error) {
	usage, err := h.Handle(ctx, GetQuotaUsageQuery{
		UserID:    userID,
		ProjectID: projectID,
	})
	if err != nil {
		return nil, err
	}
	return RunHistoryCutoff(usage.Limits.RunHistoryRetentionDays, time.Now().UTC()), nil
}
