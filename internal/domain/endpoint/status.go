package endpoint

import "fmt"

// Status is the lifecycle state of an endpoint.
type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
	StatusDeleted  Status = "deleted"
)

func (s Status) Valid() bool {
	switch s {
	case StatusActive, StatusInactive, StatusDeleted:
		return true
	default:
		return false
	}
}

func ParseStatus(value string) (Status, error) {
	s := Status(value)
	if !s.Valid() {
		return "", fmt.Errorf("invalid endpoint status %q", value)
	}
	return s, nil
}
