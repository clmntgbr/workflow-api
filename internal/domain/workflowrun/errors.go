package workflowrun

import "errors"

var (
	ErrInvalidStatusTransition = errors.New("invalid workflow run status transition")
	ErrAlreadyTerminal         = errors.New("workflow run is already terminal")
)
