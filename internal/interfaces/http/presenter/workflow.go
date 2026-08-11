package presenter

import (
	"time"

	domainworkflow "go-api/internal/domain/workflow"
)

type WorkflowDetailResponse struct {
	ID                      string    `json:"id"`
	Name                    string    `json:"name"`
	Description             string    `json:"description"`
	Status                  string    `json:"status"`
	OrganizationID          string    `json:"organizationId"`
	ScheduleIntervalMinutes int       `json:"scheduleIntervalMinutes"`
	Concurrency             int       `json:"concurrency"`
	NotificationsEnabled    bool      `json:"notificationsEnabled"`
	NotifyOnSuccess         bool      `json:"notifyOnSuccess"`
	NotifyOnFailure         bool      `json:"notifyOnFailure"`
	NotifyOnCancel          bool      `json:"notifyOnCancel"`
	CreatedAt               time.Time `json:"createdAt"`
	UpdatedAt               time.Time `json:"updatedAt"`
}

func NewWorkflowDetailResponseFromView(view domainworkflow.WorkflowView) WorkflowDetailResponse {
	return WorkflowDetailResponse{
		ID:                      view.ID.String(),
		Name:                    view.Name,
		Description:             view.Description,
		Status:                  string(view.Status),
		OrganizationID:          view.OrganizationID.String(),
		ScheduleIntervalMinutes: view.ScheduleIntervalMinutes,
		Concurrency:             view.Concurrency,
		NotificationsEnabled:    view.NotificationsEnabled,
		NotifyOnSuccess:         view.NotifyOnSuccess,
		NotifyOnFailure:         view.NotifyOnFailure,
		NotifyOnCancel:          view.NotifyOnCancel,
		CreatedAt:               view.CreatedAt,
		UpdatedAt:               view.UpdatedAt,
	}
}

func NewWorkflowDetailResponseFromEntity(w domainworkflow.Workflow) WorkflowDetailResponse {
	return WorkflowDetailResponse{
		ID:                      w.ID.String(),
		Name:                    w.Name,
		Description:             w.Description,
		Status:                  string(w.Status),
		OrganizationID:          w.OrganizationID.String(),
		ScheduleIntervalMinutes: w.ScheduleIntervalMinutes,
		Concurrency:             w.Concurrency,
		NotificationsEnabled:    w.NotificationsEnabled,
		NotifyOnSuccess:         w.NotifyOnSuccess,
		NotifyOnFailure:         w.NotifyOnFailure,
		NotifyOnCancel:          w.NotifyOnCancel,
		CreatedAt:               w.CreatedAt,
		UpdatedAt:               w.UpdatedAt,
	}
}

func NewWorkflowListResponseFromViews(views []domainworkflow.WorkflowView) []WorkflowDetailResponse {
	items := make([]WorkflowDetailResponse, 0, len(views))
	for _, view := range views {
		items = append(items, NewWorkflowDetailResponseFromView(view))
	}
	return items
}
