package steprun

import "errors"

var (
	ErrInvalidStatusTransition = errors.New("invalid step run status transition")
	ErrAlreadyTerminal         = errors.New("step run is already terminal")
	ErrNotFound                = errors.New("step run not found")
)
