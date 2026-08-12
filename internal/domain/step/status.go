package step

import "fmt"

type Status string

const (
	StatusActive  Status = "active"
	StatusDeleted Status = "deleted"
)

func (s Status) Valid() bool {
	switch s {
	case StatusActive, StatusDeleted:
		return true
	default:
		return false
	}
}

func ParseStatus(value string) (Status, error) {
	s := Status(value)
	if !s.Valid() {
		return "", fmt.Errorf("invalid step status %q", value)
	}
	return s, nil
}
