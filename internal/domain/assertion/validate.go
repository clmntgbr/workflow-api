package assertion

import (
	"regexp"
	"strconv"
	"strings"
)

func ValidateShape(source AssertionSource, path string, operator AssertionOperator, expectedValue string) error {
	if !source.Valid() {
		return ErrInvalidSource
	}
	if !operator.Valid() {
		return ErrInvalidOperator
	}

	trimmedPath := strings.TrimSpace(path)
	trimmedExpected := strings.TrimSpace(expectedValue)

	switch source {
	case SourceStatus:
		if trimmedPath != "" {
			return ErrPathForbidden
		}
	case SourceHeader, SourceBody:
		if trimmedPath == "" {
			return ErrPathRequired
		}
	}

	if operator.RequiresExpectedValue() {
		if trimmedExpected == "" {
			return ErrExpectedValueRequired
		}
	} else if trimmedExpected != "" {
		return ErrExpectedValueForbidden
	}

	switch operator {
	case OperatorGreaterThan, OperatorLessThan:
		if _, err := strconv.ParseFloat(trimmedExpected, 64); err != nil {
			return ErrInvalidExpectedNumber
		}
	case OperatorMatchesRegex:
		if _, err := regexp.Compile(trimmedExpected); err != nil {
			return ErrInvalidExpectedRegex
		}
	}

	return nil
}
