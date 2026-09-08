package workflowrun

import "errors"

var (
	ErrInvalidStatusTransition = errors.New("invalid workflow run status transition")
	ErrAlreadyTerminal         = errors.New("workflow run is already terminal")
	ErrAlreadyInProgress       = errors.New("workflow run already in progress")
	ErrNoRunInProgress         = errors.New("no workflow run in progress")
	ErrWorkflowNotFound        = errors.New("workflow not found")
	ErrFromStepNotFound        = errors.New("from step not found")
	ErrMissingPreviousStepRun  = errors.New("a previous step has no successful run")
)
