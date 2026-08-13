package workflowrun

import "fmt"

type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusSuccess   Status = "success"
	StatusFailed    Status = "failed"
	StatusCancelled Status = "cancelled"
)

func (s Status) Valid() bool {
	switch s {
	case StatusPending, StatusRunning, StatusSuccess, StatusFailed, StatusCancelled:
		return true
	default:
		return false
	}
}

func (s Status) IsTerminal() bool {
	switch s {
	case StatusSuccess, StatusFailed, StatusCancelled:
		return true
	default:
		return false
	}
}

func ParseStatus(value string) (Status, error) {
	s := Status(value)
	if !s.Valid() {
		return "", fmt.Errorf("invalid workflow run status %q", value)
	}
	return s, nil
}
