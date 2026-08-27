package presenter

import (
	"time"

	domainworkflow "go-api/internal/domain/workflow"
)

type WorkflowListResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Status      string  `json:"status"`
}

type WorkflowDetailResponse struct {
	WorkflowListResponse
	ScheduleType          string     `json:"scheduleType"`
	ScheduleIntervalValue int        `json:"scheduleIntervalValue"`
	ScheduleIntervalUnit  *string    `json:"scheduleIntervalUnit"`
	ScheduleAt            *time.Time `json:"scheduleAt"`
	ScheduleTimezone      string     `json:"scheduleTimezone"`
	NextRunAt             *time.Time `json:"nextRunAt"`
	NotificationsEnabled  bool       `json:"notificationsEnabled"`
	NotifyOnSuccess       bool       `json:"notifyOnSuccess"`
	NotifyOnFailure       bool       `json:"notifyOnFailure"`
	NotifyOnCancel        bool       `json:"notifyOnCancel"`
}

func NewWorkflowListResponseFromView(view domainworkflow.WorkflowView) WorkflowListResponse {
	return WorkflowListResponse{
		ID:          view.ID.String(),
		Name:        view.Name,
		Description: optionalNonEmptyString(view.Description),
		Status:      string(view.Status),
	}
}

func NewWorkflowListResponseFromViews(views []domainworkflow.WorkflowView) []WorkflowListResponse {
	items := make([]WorkflowListResponse, 0, len(views))
	for _, view := range views {
		items = append(items, NewWorkflowListResponseFromView(view))
	}
	return items
}

func NewWorkflowDetailResponseFromView(view domainworkflow.WorkflowView) WorkflowDetailResponse {
	return workflowDetailResponse(
		NewWorkflowListResponseFromView(view),
		string(view.ScheduleType),
		view.ScheduleIntervalValue,
		view.ScheduleIntervalUnit,
		view.ScheduleAt,
		view.ScheduleTimezone,
		view.NextRunAt,
		view.NotificationsEnabled,
		view.NotifyOnSuccess,
		view.NotifyOnFailure,
		view.NotifyOnCancel,
	)
}

func NewWorkflowDetailResponseFromEntity(w domainworkflow.Workflow) WorkflowDetailResponse {
	return workflowDetailResponse(
		WorkflowListResponse{
			ID:          w.ID.String(),
			Name:        w.Name,
			Description: optionalNonEmptyString(w.Description),
			Status:      string(w.Status),
		},
		string(w.ScheduleType),
		w.ScheduleIntervalValue,
		w.ScheduleIntervalUnit,
		w.ScheduleAt,
		w.ScheduleTimezone,
		w.NextRunAt,
		w.NotificationsEnabled,
		w.NotifyOnSuccess,
		w.NotifyOnFailure,
		w.NotifyOnCancel,
	)
}

func workflowDetailResponse(
	list WorkflowListResponse,
	scheduleType string,
	scheduleIntervalValue int,
	scheduleIntervalUnit domainworkflow.ScheduleUnit,
	scheduleAt *time.Time,
	scheduleTimezone string,
	nextRunAt *time.Time,
	notificationsEnabled bool,
	notifyOnSuccess bool,
	notifyOnFailure bool,
	notifyOnCancel bool,
) WorkflowDetailResponse {
	if scheduleType == "" {
		scheduleType = string(domainworkflow.ScheduleTypeNone)
	}
	if scheduleTimezone == "" {
		scheduleTimezone = "UTC"
	}
	return WorkflowDetailResponse{
		WorkflowListResponse:  list,
		ScheduleType:          scheduleType,
		ScheduleIntervalValue: scheduleIntervalValue,
		ScheduleIntervalUnit:  optionalNonEmptyString(string(scheduleIntervalUnit)),
		ScheduleAt:            scheduleAt,
		ScheduleTimezone:      scheduleTimezone,
		NextRunAt:             nextRunAt,
		NotificationsEnabled:  notificationsEnabled,
		NotifyOnSuccess:       notifyOnSuccess,
		NotifyOnFailure:       notifyOnFailure,
		NotifyOnCancel:        notifyOnCancel,
	}
}

func optionalNonEmptyString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
