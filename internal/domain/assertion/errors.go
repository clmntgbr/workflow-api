package assertion

import "errors"

var (
	ErrNotFound               = errors.New("assertion not found")
	ErrInvalidSource          = errors.New("invalid assertion source")
	ErrInvalidOperator        = errors.New("invalid assertion operator")
	ErrPathRequired           = errors.New("path is required")
	ErrPathForbidden          = errors.New("path must be empty for status assertions")
	ErrExpectedValueRequired  = errors.New("expectedValue is required for this operator")
	ErrExpectedValueForbidden = errors.New("expectedValue must be empty for this operator")
	ErrInvalidExpectedNumber  = errors.New("expectedValue must be a valid number")
	ErrInvalidExpectedRegex   = errors.New("expectedValue must be a valid regular expression")
)
