package assertion

import "fmt"

type AssertionSource string

const (
	SourceStatus AssertionSource = "status"
	SourceHeader AssertionSource = "header"
	SourceBody   AssertionSource = "body"
)

func (s AssertionSource) Valid() bool {
	switch s {
	case SourceStatus, SourceHeader, SourceBody:
		return true
	default:
		return false
	}
}

func ParseSource(raw string) (AssertionSource, error) {
	s := AssertionSource(raw)
	if !s.Valid() {
		return "", fmt.Errorf("%w: %s", ErrInvalidSource, raw)
	}
	return s, nil
}
