package workflow

import "errors"

var (
	ErrScheduleIntervalTooShort = errors.New("schedule interval must be at least 1 minute")
	ErrInvalidSchedule          = errors.New("invalid workflow schedule")
	ErrInvalidScheduleTimezone  = errors.New("invalid schedule timezone")
	ErrInvalidStatusTransition  = errors.New("invalid workflow status transition")
)
