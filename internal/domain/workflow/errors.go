package workflow

import "errors"

var (
	ErrScheduleIntervalTooShort = errors.New("schedule interval must be at least 10 minutes")
	ErrInvalidSchedule          = errors.New("invalid workflow schedule")
	ErrInvalidScheduleTimezone  = errors.New("invalid schedule timezone")
)
