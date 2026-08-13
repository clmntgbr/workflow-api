package presenter

import (
	"time"

	domainworkflow "go-api/internal/domain/workflow"
)

type WorkflowDetailResponse struct {
	ID                    string     `json:"id"`
	Name                  string     `json:"name"`
	Description           *string    `json:"description"`
	Status                string     `json:"status"`
	OrganizationID        string     `json:"organizationId"`
	ScheduleType          string     `json:"scheduleType"`
	ScheduleIntervalValue int        `json:"scheduleIntervalValue"`
	ScheduleIntervalUnit  *string    `json:"scheduleIntervalUnit"`
	ScheduleAt            *time.Time `json:"scheduleAt"`
	ScheduleTimezone      string     `json:"scheduleTimezone"`
	NextRunAt             *time.Time `json:"nextRunAt"`
	Concurrency           int        `json:"concurrency"`
	NotificationsEnabled  bool       `json:"notificationsEnabled"`
	NotifyOnSuccess       bool       `json:"notifyOnSuccess"`
	NotifyOnFailure       bool       `json:"notifyOnFailure"`
	NotifyOnCancel        bool       `json:"notifyOnCancel"`
	CreatedAt             time.Time  `json:"createdAt"`
	UpdatedAt             time.Time  `json:"updatedAt"`
}

func NewWorkflowDetailResponseFromView(view domainworkflow.WorkflowView) WorkflowDetailResponse {
	return workflowDetailResponse(
		view.ID.String(),
		view.Name,
		view.Description,
		string(view.Status),
		view.OrganizationID.String(),
		string(view.ScheduleType),
		view.ScheduleIntervalValue,
		view.ScheduleIntervalUnit,
		view.ScheduleAt,
		view.ScheduleTimezone,
		view.NextRunAt,
		view.Concurrency,
		view.NotificationsEnabled,
		view.NotifyOnSuccess,
		view.NotifyOnFailure,
		view.NotifyOnCancel,
		view.CreatedAt,
		view.UpdatedAt,
	)
}

func NewWorkflowDetailResponseFromEntity(w domainworkflow.Workflow) WorkflowDetailResponse {
	return workflowDetailResponse(
		w.ID.String(),
		w.Name,
		w.Description,
		string(w.Status),
		w.OrganizationID.String(),
		string(w.ScheduleType),
		w.ScheduleIntervalValue,
		w.ScheduleIntervalUnit,
		w.ScheduleAt,
		w.ScheduleTimezone,
		w.NextRunAt,
		w.Concurrency,
		w.NotificationsEnabled,
		w.NotifyOnSuccess,
		w.NotifyOnFailure,
		w.NotifyOnCancel,
		w.CreatedAt,
		w.UpdatedAt,
	)
}

func NewWorkflowListResponseFromViews(views []domainworkflow.WorkflowView) []WorkflowDetailResponse {
	items := make([]WorkflowDetailResponse, 0, len(views))
	for _, view := range views {
		items = append(items, NewWorkflowDetailResponseFromView(view))
	}
	return items
}

func workflowDetailResponse(
	id string,
	name string,
	description string,
	status string,
	organizationID string,
	scheduleType string,
	scheduleIntervalValue int,
	scheduleIntervalUnit domainworkflow.ScheduleUnit,
	scheduleAt *time.Time,
	scheduleTimezone string,
	nextRunAt *time.Time,
	concurrency int,
	notificationsEnabled bool,
	notifyOnSuccess bool,
	notifyOnFailure bool,
	notifyOnCancel bool,
	createdAt time.Time,
	updatedAt time.Time,
) WorkflowDetailResponse {
	if scheduleType == "" {
		scheduleType = string(domainworkflow.ScheduleTypeNone)
	}
	if scheduleTimezone == "" {
		scheduleTimezone = "UTC"
	}
	return WorkflowDetailResponse{
		ID:                    id,
		Name:                  name,
		Description:           optionalNonEmptyString(description),
		Status:                status,
		OrganizationID:        organizationID,
		ScheduleType:          scheduleType,
		ScheduleIntervalValue: scheduleIntervalValue,
		ScheduleIntervalUnit:  optionalNonEmptyString(string(scheduleIntervalUnit)),
		ScheduleAt:            scheduleAt,
		ScheduleTimezone:      scheduleTimezone,
		NextRunAt:             nextRunAt,
		Concurrency:           concurrency,
		NotificationsEnabled:  notificationsEnabled,
		NotifyOnSuccess:       notifyOnSuccess,
		NotifyOnFailure:       notifyOnFailure,
		NotifyOnCancel:        notifyOnCancel,
		CreatedAt:             createdAt,
		UpdatedAt:             updatedAt,
	}
}

func optionalNonEmptyString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
