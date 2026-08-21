package variable

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrInvalidKind      = errors.New("invalid variable kind")
	ErrStepRequired     = errors.New("stepId is required for extracted variables")
	ErrStepForbidden    = errors.New("stepId must be empty for static variables")
	ErrPathRequired     = errors.New("path is required for extracted variables")
	ErrPathForbidden    = errors.New("path must be empty for static variables")
	ErrValueRequired    = errors.New("value is required for static variables")
	ErrValueForbidden   = errors.New("value must be empty for extracted variables")
)

func ValidateShape(kind Kind, stepID *uuid.UUID, path string, value any) error {
	if !kind.Valid() {
		return ErrInvalidKind
	}
	trimmedPath := strings.TrimSpace(path)

	switch kind {
	case KindExtracted:
		if stepID == nil || *stepID == uuid.Nil {
			return ErrStepRequired
		}
		if trimmedPath == "" {
			return ErrPathRequired
		}
		if value != nil {
			return ErrValueForbidden
		}
	case KindStatic:
		if stepID != nil && *stepID != uuid.Nil {
			return ErrStepForbidden
		}
		if trimmedPath != "" {
			return ErrPathForbidden
		}
		if value == nil {
			return ErrValueRequired
		}
	}
	return nil
}

func ValidateIdentity(name, key string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("name is required")
	}
	if strings.TrimSpace(key) == "" {
		return errors.New("key is required")
	}
	return nil
}
