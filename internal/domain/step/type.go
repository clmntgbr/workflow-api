package step

import (
	"errors"
	"strings"

	domaincondition "go-api/internal/domain/condition"
)

type Type string

const (
	TypeHTTP       Type = "http"
	TypeDelay      Type = "delay"
	TypeCondition  Type = "condition"
)

func (t Type) Valid() bool {
	switch t {
	case TypeHTTP, TypeDelay, TypeCondition:
		return true
	default:
		return false
	}
}

func ParseType(value string) (Type, error) {
	t := Type(value)
	if !t.Valid() {
		return "", errors.New("invalid step type")
	}
	return t, nil
}

var (
	ErrInvalidStepTypeConfig      = errors.New("invalid step type configuration")
	ErrNonHTTPStepCannotHaveExtras = errors.New("only HTTP steps can have variables or assertions")
	// ErrDelayStepCannotHaveExtras is kept for backward compatibility with existing callers.
	ErrDelayStepCannotHaveExtras = ErrNonHTTPStepCannotHaveExtras
)

func ValidateConfig(s *Step) error {
	if s == nil {
		return ErrInvalidStepTypeConfig
	}
	if !s.Type.Valid() {
		return ErrInvalidStepTypeConfig
	}

	switch s.Type {
	case TypeHTTP:
		if s.EndpointID == nil {
			return ErrInvalidStepTypeConfig
		}
		if s.DelayDurationSeconds != 0 {
			return ErrInvalidStepTypeConfig
		}
		if s.Expression != nil {
			return ErrInvalidStepTypeConfig
		}
	case TypeDelay:
		if s.EndpointID != nil {
			return ErrInvalidStepTypeConfig
		}
		if s.DelayDurationSeconds <= 0 {
			return ErrInvalidStepTypeConfig
		}
		if s.Expression != nil {
			return ErrInvalidStepTypeConfig
		}
	case TypeCondition:
		if s.EndpointID != nil {
			return ErrInvalidStepTypeConfig
		}
		if s.DelayDurationSeconds != 0 {
			return ErrInvalidStepTypeConfig
		}
		if s.Expression == nil || strings.TrimSpace(*s.Expression) == "" {
			return ErrInvalidStepTypeConfig
		}
		if err := domaincondition.ValidateSyntax(strings.TrimSpace(*s.Expression)); err != nil {
			return err
		}
	default:
		return ErrInvalidStepTypeConfig
	}
	return nil
}
