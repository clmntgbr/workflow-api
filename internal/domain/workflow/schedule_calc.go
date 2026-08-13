package workflow

import "time"

const minScheduleIntervalMinutes = 10

func (w *Workflow) validateSchedule() error {
	if !w.ScheduleType.Valid() {
		return ErrInvalidSchedule
	}
	if w.ScheduleTimezone != "" {
		if _, err := time.LoadLocation(w.ScheduleTimezone); err != nil {
			return ErrInvalidScheduleTimezone
		}
	}

	switch w.ScheduleType {
	case ScheduleTypeNone:
		return nil
	case ScheduleTypeRecurring:
		if w.ScheduleIntervalValue <= 0 || !w.ScheduleIntervalUnit.Valid() {
			return ErrInvalidSchedule
		}
		return w.validateScheduleInterval()
	case ScheduleTypeOnce:
		if w.ScheduleAt == nil {
			return ErrInvalidSchedule
		}
		return nil
	default:
		return ErrInvalidSchedule
	}
}

func (w *Workflow) validateScheduleInterval() error {
	if w.ScheduleType != ScheduleTypeRecurring {
		return nil
	}
	if toMinutes(w.ScheduleIntervalValue, w.ScheduleIntervalUnit) < minScheduleIntervalMinutes {
		return ErrScheduleIntervalTooShort
	}
	return nil
}

func (w *Workflow) RecalculateNextRunAt(now time.Time) {
	switch w.ScheduleType {
	case ScheduleTypeNone:
		w.NextRunAt = nil
	case ScheduleTypeRecurring:
		next := addInterval(now.In(w.scheduleLocation()), w.ScheduleIntervalValue, w.ScheduleIntervalUnit)
		w.NextRunAt = &next
	case ScheduleTypeOnce:
		if w.ScheduleAt != nil && w.ScheduleAt.After(now) {
			at := w.ScheduleAt.UTC()
			w.NextRunAt = &at
		} else {
			w.NextRunAt = nil
		}
	default:
		w.NextRunAt = nil
	}
}

// AdvanceAfterScheduledStart moves the schedule forward after the scheduler
// successfully started a run (strict cadence from the previous NextRunAt).
func (w *Workflow) AdvanceAfterScheduledStart(now time.Time) {
	switch w.ScheduleType {
	case ScheduleTypeRecurring:
		base := now
		if w.NextRunAt != nil {
			base = *w.NextRunAt
		}
		next := addInterval(base.In(w.scheduleLocation()), w.ScheduleIntervalValue, w.ScheduleIntervalUnit)
		for !next.After(now) {
			next = addInterval(next, w.ScheduleIntervalValue, w.ScheduleIntervalUnit)
		}
		w.NextRunAt = &next
		w.UpdatedAt = now
	case ScheduleTypeOnce:
		w.ScheduleType = ScheduleTypeNone
		w.ScheduleAt = nil
		w.NextRunAt = nil
		w.UpdatedAt = now
	}
}

func (w *Workflow) scheduleLocation() *time.Location {
	if w.ScheduleTimezone == "" {
		return time.UTC
	}
	loc, err := time.LoadLocation(w.ScheduleTimezone)
	if err != nil {
		return time.UTC
	}
	return loc
}

func addInterval(t time.Time, value int, unit ScheduleUnit) time.Time {
	switch unit {
	case ScheduleUnitMinute:
		return t.Add(time.Duration(value) * time.Minute)
	case ScheduleUnitHour:
		return t.Add(time.Duration(value) * time.Hour)
	case ScheduleUnitDay:
		return t.AddDate(0, 0, value)
	case ScheduleUnitWeek:
		return t.AddDate(0, 0, value*7)
	case ScheduleUnitMonth:
		return t.AddDate(0, value, 0)
	case ScheduleUnitYear:
		return t.AddDate(value, 0, 0)
	default:
		return t
	}
}

func toMinutes(value int, unit ScheduleUnit) int {
	if value <= 0 {
		return 0
	}
	switch unit {
	case ScheduleUnitMinute:
		return value
	case ScheduleUnitHour:
		return value * 60
	case ScheduleUnitDay:
		return value * 60 * 24
	case ScheduleUnitWeek:
		return value * 60 * 24 * 7
	case ScheduleUnitMonth:
		return value * 60 * 24 * 30
	case ScheduleUnitYear:
		return value * 60 * 24 * 365
	default:
		return 0
	}
}
