package dto

type CreateWorkflowRequest struct {
	Name                    string `json:"name"`
	Description             string `json:"description"`
	OrganizationID          string `json:"organizationId"`
	ScheduleIntervalMinutes *int   `json:"scheduleIntervalMinutes"`
	Concurrency             *int   `json:"concurrency"`
	NotificationsEnabled    *bool  `json:"notificationsEnabled"`
	NotifyOnSuccess         *bool  `json:"notifyOnSuccess"`
	NotifyOnFailure         *bool  `json:"notifyOnFailure"`
	NotifyOnCancel          *bool  `json:"notifyOnCancel"`
}

type UpdateWorkflowRequest struct {
	Name                    string `json:"name"`
	Description             string `json:"description"`
	Status                  string `json:"status"`
	ScheduleIntervalMinutes *int   `json:"scheduleIntervalMinutes"`
	Concurrency             *int   `json:"concurrency"`
	NotificationsEnabled    *bool  `json:"notificationsEnabled"`
	NotifyOnSuccess         *bool  `json:"notifyOnSuccess"`
	NotifyOnFailure         *bool  `json:"notifyOnFailure"`
	NotifyOnCancel          *bool  `json:"notifyOnCancel"`
}
