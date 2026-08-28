package step

import "errors"

type Type string

const (
	TypeHTTP  Type = "http"
	TypeDelay Type = "delay"
)

func (t Type) Valid() bool {
	switch t {
	case TypeHTTP, TypeDelay:
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
	ErrInvalidStepTypeConfig     = errors.New("invalid step type configuration")
	ErrDelayStepCannotHaveExtras = errors.New("delay steps cannot have variables or assertions")
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
	case TypeDelay:
		if s.EndpointID != nil {
			return ErrInvalidStepTypeConfig
		}
		if s.DelayDurationSeconds <= 0 {
			return ErrInvalidStepTypeConfig
		}
	default:
		return ErrInvalidStepTypeConfig
	}
	return nil
}
