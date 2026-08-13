package workflow

import "fmt"

type ScheduleType string

const (
	ScheduleTypeNone      ScheduleType = "none"
	ScheduleTypeRecurring ScheduleType = "recurring"
	ScheduleTypeOnce      ScheduleType = "once"
)

func (t ScheduleType) Valid() bool {
	switch t {
	case ScheduleTypeNone, ScheduleTypeRecurring, ScheduleTypeOnce:
		return true
	default:
		return false
	}
}

func ParseScheduleType(value string) (ScheduleType, error) {
	if value == "" {
		return ScheduleTypeNone, nil
	}
	t := ScheduleType(value)
	if !t.Valid() {
		return "", fmt.Errorf("invalid schedule type %q", value)
	}
	return t, nil
}

type ScheduleUnit string

const (
	ScheduleUnitMinute ScheduleUnit = "minute"
	ScheduleUnitHour   ScheduleUnit = "hour"
	ScheduleUnitDay    ScheduleUnit = "day"
	ScheduleUnitWeek   ScheduleUnit = "week"
	ScheduleUnitMonth  ScheduleUnit = "month"
	ScheduleUnitYear   ScheduleUnit = "year"
)

func (u ScheduleUnit) Valid() bool {
	switch u {
	case ScheduleUnitMinute, ScheduleUnitHour, ScheduleUnitDay, ScheduleUnitWeek, ScheduleUnitMonth, ScheduleUnitYear:
		return true
	default:
		return false
	}
}

func ParseScheduleUnit(value string) (ScheduleUnit, error) {
	if value == "" {
		return "", nil
	}
	u := ScheduleUnit(value)
	if !u.Valid() {
		return "", fmt.Errorf("invalid schedule unit %q", value)
	}
	return u, nil
}
