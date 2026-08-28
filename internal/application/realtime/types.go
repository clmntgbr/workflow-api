package realtime

const (
	ActionCreated                   = "created"
	ActionUpdated                   = "updated"
	ActionDeleted                   = "deleted"
	ActionActiveProjectChanged = "active_project_changed"
	ActionMemberAdded               = "member_added"
	ActionMemberRemoved             = "member_removed"
	ActionStarted                   = "started"
	ActionSucceeded                 = "succeeded"
	ActionFailed                    = "failed"
	ActionCancelled                 = "cancelled"
	ActionFinished                  = "finished"
	ActionImported                  = "imported"

	EntityUser         = "user"
	EntityProject      = "project"
	EntityWorkflow     = "workflow"
	EntityEndpoint     = "endpoint"
	EntityStep         = "step"
	EntityConnection   = "connection"
	EntityVariable     = "variable"
	EntityAssertion    = "assertion"
	EntityWorkflowRun  = "workflowRun"
	EntityStepRun      = "stepRun"
)

func EventType(entity, action string) string {
	return entity + "." + action
}
