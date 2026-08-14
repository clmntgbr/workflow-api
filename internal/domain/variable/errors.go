package variable

import "errors"

var (
	ErrNotFound     = errors.New("variable not found")
	ErrDuplicateKey = errors.New("variable key already exists in workflow")
	ErrInvalidPath  = errors.New("invalid variable path")
	ErrInvalidRef   = errors.New("variable reference is invalid")
	ErrNotAncestor  = errors.New("variable step is not an ancestor of the referencing step")
	ErrMissingValue = errors.New("variable value not found in context")
)
