package activitylog

const (
	LevelInfo    = "info"
	LevelWarning = "warning"
	LevelError   = "error"
)

const (
	SubjectWorkflow    = "workflow"
	SubjectStep        = "step"
	SubjectConnection  = "connection"
	SubjectVariable    = "variable"
	SubjectAssertion   = "assertion"
	SubjectEndpoint    = "endpoint"
	SubjectWorkflowRun = "workflow_run"
	SubjectStepRun     = "step_run"
)

const (
	ActorUser     = "user"
	ActorSchedule = "schedule"
	ActorSystem   = "system"
	ActorAPI      = "api"
	ActorCLI      = "cli"
	ActorWebhook  = "webhook"
)

const (
	ActionWorkflowCreated          = "workflow.created"
	ActionWorkflowUpdated          = "workflow.updated"
	ActionWorkflowDeleted          = "workflow.deleted"
	ActionStepCreated              = "step.created"
	ActionStepUpdated              = "step.updated"
	ActionStepDeleted              = "step.deleted"
	ActionConnectionCreated        = "connection.created"
	ActionConnectionDeleted        = "connection.deleted"
	ActionVariableCreated          = "variable.created"
	ActionVariableUpdated          = "variable.updated"
	ActionAssertionCreated         = "assertion.created"
	ActionAssertionUpdated         = "assertion.updated"
	ActionEndpointCreated          = "endpoint.created"
	ActionEndpointUpdated          = "endpoint.updated"
	ActionEndpointDeleted          = "endpoint.deleted"
	ActionEndpointImported         = "endpoint.imported"
	ActionWorkflowRunStarted       = "workflow_run.started"
	ActionWorkflowRunFinished      = "workflow_run.finished"
	ActionWorkflowRunCancelled     = "workflow_run.cancelled"
	ActionWorkflowRunScheduledSkip = "workflow_run.scheduled_skipped"
	ActionStepRunQueued            = "step_run.queued"
	ActionStepRunStarted           = "step_run.started"
	ActionStepRunSucceeded         = "step_run.succeeded"
	ActionStepRunFailed            = "step_run.failed"
)
