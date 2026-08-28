package presenter

import (
	"time"

	domainactivitylog "go-api/internal/domain/activitylog"
	"go-api/internal/domain/paginate"
)

type ActivityLogResponse struct {
	ID              string         `json:"id"`
	ProjectID       string         `json:"projectId"`
	Action          string         `json:"action"`
	SubjectType     string         `json:"subjectType"`
	SubjectID       string         `json:"subjectId"`
	WorkflowID      *string        `json:"workflowId"`
	WorkflowRunID   *string        `json:"workflowRunId"`
	StepID          *string        `json:"stepId"`
	StepRunID       *string        `json:"stepRunId"`
	ActorType       *string        `json:"actorType"`
	ActorUserID     *string        `json:"actorUserId"`
	Level           string         `json:"level"`
	Message         string  `json:"message"`
	SourceEventID   string  `json:"sourceEventId"`
	SourceEventType string  `json:"sourceEventType"`
	OccurredAt      time.Time      `json:"occurredAt"`
	CreatedAt       time.Time      `json:"createdAt"`
}

func NewActivityLogResponse(view domainactivitylog.View) ActivityLogResponse {
	return ActivityLogResponse{
		ID:              view.ID.String(),
		ProjectID:       view.ProjectID.String(),
		Action:          view.Action,
		SubjectType:     view.SubjectType,
		SubjectID:       view.SubjectID.String(),
		WorkflowID:      optionalUUIDString(view.WorkflowID),
		WorkflowRunID:   optionalUUIDString(view.WorkflowRunID),
		StepID:          optionalUUIDString(view.StepID),
		StepRunID:       optionalUUIDString(view.StepRunID),
		ActorType:       optionalNonEmptyString(view.ActorType),
		ActorUserID:     optionalUUIDString(view.ActorUserID),
		Level:           view.Level,
		Message:         view.Message,
		SourceEventID:   view.SourceEventID.String(),
		SourceEventType: view.SourceEventType,
		OccurredAt:      view.OccurredAt,
		CreatedAt:       view.CreatedAt,
	}
}

func NewActivityLogListResponse(views []domainactivitylog.View) []ActivityLogResponse {
	items := make([]ActivityLogResponse, 0, len(views))
	for _, view := range views {
		items = append(items, NewActivityLogResponse(view))
	}
	return items
}

func NewActivityLogPaginateResponse(
	views []domainactivitylog.View,
	total int64,
	query paginate.PaginateQuery,
) paginate.PaginateResponse {
	return paginate.NewPaginateResponse(
		NewActivityLogListResponse(views),
		int(total),
		query,
	)
}
