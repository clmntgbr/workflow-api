package workflow

import "fmt"

type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
	StatusDeleted  Status = "deleted"
	StatusCanceled Status = "canceled"
)

func (s Status) Valid() bool {
	switch s {
	case StatusActive, StatusInactive, StatusDeleted, StatusCanceled:
		return true
	default:
		return false
	}
}

func ParseStatus(value string) (Status, error) {
	s := Status(value)
	if !s.Valid() {
		return "", fmt.Errorf("invalid workflow status %q", value)
	}
	return s, nil
}
