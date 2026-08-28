package steprun

import "fmt"

type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusSuccess   Status = "success"
	StatusFailed    Status = "failed"
	StatusCancelled Status = "cancelled"
	StatusWaiting   Status = "waiting"
	StatusSkipped   Status = "skipped"
)

func (s Status) Valid() bool {
	switch s {
	case StatusPending, StatusRunning, StatusWaiting, StatusSuccess, StatusFailed, StatusCancelled, StatusSkipped:
		return true
	default:
		return false
	}
}

func (s Status) IsTerminal() bool {
	switch s {
	case StatusSuccess, StatusFailed, StatusCancelled, StatusSkipped:
		return true
	default:
		return false
	}
}

func ParseStatus(value string) (Status, error) {
	s := Status(value)
	if !s.Valid() {
		return "", fmt.Errorf("invalid step run status %q", value)
	}
	return s, nil
}
