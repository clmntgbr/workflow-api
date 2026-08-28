package activitylog

import (
	domainactivitylog "go-api/internal/domain/activitylog"

	"github.com/google/uuid"
)

type messageHints struct {
	WorkflowName      string
	WorkflowStatus    string
	StepName          string
	VariableKey       string
	VariableKind      string
	EndpointName      string
	Method            string
	URL               string
	Attempt           int
	MaxAttempts       int
	StatusCode        int
	Error             string
	FinishType        string
	TriggeredBy       string
	SkipReason        string
	ImportCount       int
	AssertionSource   string
	AssertionOperator string
	ExpectedValue     string
	ActorUserName     string
	SourceStepID      uuid.UUID
	TargetStepID      uuid.UUID
	SourceStepName    string
	TargetStepName    string
	PositionX         float64
	PositionY         float64
}

type projectedLog struct {
	entry *domainactivitylog.Entry
	hints messageHints
}
