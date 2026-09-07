package runexport

import "errors"

var (
	ErrNotFound         = errors.New("run export not found")
	ErrInvalidDateRange = errors.New("invalid export date range")
	ErrFileTooLarge     = errors.New("export file is too large to send by email")
)
