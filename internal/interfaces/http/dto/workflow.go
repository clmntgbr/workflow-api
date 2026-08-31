package dto

import "time"

type CreateWorkflowRequest struct {
	Name                  string     `json:"name" validate:"required,min=1,max=255"`
	Description           string     `json:"description" validate:"omitempty,max=2000"`
	ScheduleType          string     `json:"scheduleType" validate:"omitempty,schedule_type"`
	ScheduleIntervalValue *int       `json:"scheduleIntervalValue" validate:"omitempty,min=0"`
	ScheduleIntervalUnit  string     `json:"scheduleIntervalUnit" validate:"omitempty,schedule_unit"`
	ScheduleAt            *time.Time `json:"scheduleAt" validate:"omitempty"`
	ScheduleTimezone      string     `json:"scheduleTimezone" validate:"omitempty,max=64"`
	Concurrency           *int       `json:"concurrency" validate:"omitempty,min=1,max=100"`
	NotificationsEnabled  *bool      `json:"notificationsEnabled" validate:"omitempty"`
	NotifyOnSuccess       *bool      `json:"notifyOnSuccess" validate:"omitempty"`
	NotifyOnFailure       *bool      `json:"notifyOnFailure" validate:"omitempty"`
	NotifyOnCancel        *bool      `json:"notifyOnCancel" validate:"omitempty"`
}

type UpdateWorkflowRequest struct {
	Name                  string     `json:"name" validate:"required,min=1,max=255"`
	Description           string     `json:"description" validate:"omitempty,max=2000"`
	Status                string     `json:"status" validate:"required,workflow_status"`
	ScheduleType          string     `json:"scheduleType" validate:"required,schedule_type"`
	ScheduleIntervalValue *int       `json:"scheduleIntervalValue" validate:"omitempty,min=0"`
	ScheduleIntervalUnit  string     `json:"scheduleIntervalUnit" validate:"omitempty,schedule_unit"`
	ScheduleAt            *time.Time `json:"scheduleAt" validate:"omitempty"`
	ScheduleTimezone      string     `json:"scheduleTimezone" validate:"omitempty,max=64"`
	Concurrency           *int       `json:"concurrency" validate:"omitempty,min=1,max=100"`
	NotificationsEnabled  *bool      `json:"notificationsEnabled" validate:"omitempty"`
	NotifyOnSuccess       *bool      `json:"notifyOnSuccess" validate:"omitempty"`
	NotifyOnFailure       *bool      `json:"notifyOnFailure" validate:"omitempty"`
	NotifyOnCancel        *bool      `json:"notifyOnCancel" validate:"omitempty"`
}
