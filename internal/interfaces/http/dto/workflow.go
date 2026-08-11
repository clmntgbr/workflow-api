package dto

type CreateWorkflowRequest struct {
	Name                    string `json:"name" validate:"required,min=1,max=255"`
	Description             string `json:"description" validate:"omitempty,max=2000"`
	ScheduleIntervalMinutes *int   `json:"scheduleIntervalMinutes" validate:"omitempty,min=0"`
	Concurrency             *int   `json:"concurrency" validate:"omitempty,min=1,max=100"`
	NotificationsEnabled    *bool  `json:"notificationsEnabled" validate:"omitempty"`
	NotifyOnSuccess         *bool  `json:"notifyOnSuccess" validate:"omitempty"`
	NotifyOnFailure         *bool  `json:"notifyOnFailure" validate:"omitempty"`
	NotifyOnCancel          *bool  `json:"notifyOnCancel" validate:"omitempty"`
}

type UpdateWorkflowRequest struct {
	Name                    string `json:"name" validate:"required,min=1,max=255"`
	Description             string `json:"description" validate:"omitempty,max=2000"`
	Status                  string `json:"status" validate:"required,workflow_status"`
	ScheduleIntervalMinutes *int   `json:"scheduleIntervalMinutes" validate:"required,min=0"`
	Concurrency             *int   `json:"concurrency" validate:"required,min=1,max=100"`
	NotificationsEnabled    *bool  `json:"notificationsEnabled" validate:"required"`
	NotifyOnSuccess         *bool  `json:"notifyOnSuccess" validate:"required"`
	NotifyOnFailure         *bool  `json:"notifyOnFailure" validate:"required"`
	NotifyOnCancel          *bool  `json:"notifyOnCancel" validate:"required"`
}
