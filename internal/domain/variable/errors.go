package variable

import "errors"

var (
	ErrNotFound     = errors.New("variable not found")
	ErrDuplicateKey = errors.New("variable key already exists in workflow")
	ErrInvalidPath  = errors.New("invalid variable path")
	ErrMissingValue = errors.New("variable value not found in context")
)
