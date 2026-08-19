package workflowrun

import "errors"

var (
	ErrInvalidStatusTransition = errors.New("invalid workflow run status transition")
	ErrAlreadyTerminal         = errors.New("workflow run is already terminal")
	ErrAlreadyInProgress       = errors.New("workflow run already in progress")
	ErrNoRunInProgress         = errors.New("no workflow run in progress")
	ErrWorkflowNotFound        = errors.New("workflow not found")
)
