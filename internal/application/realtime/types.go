package realtime

const (
	ActionCreated                   = "created"
	ActionUpdated                   = "updated"
	ActionDeleted                   = "deleted"
	ActionActiveOrganizationChanged = "active_organization_changed"
	ActionMemberAdded               = "member_added"
	ActionMemberRemoved             = "member_removed"
	ActionStarted                   = "started"
	ActionSucceeded                 = "succeeded"
	ActionFailed                    = "failed"
	ActionCancelled                 = "cancelled"
	ActionImported                  = "imported"

	EntityUser         = "user"
	EntityOrganization = "organization"
	EntityWorkflow     = "workflow"
	EntityEndpoint     = "endpoint"
	EntityStep         = "step"
	EntityConnection   = "connection"
	EntityVariable     = "variable"
	EntityWorkflowRun  = "workflowRun"
	EntityStepRun      = "stepRun"
)

func EventType(entity, action string) string {
	return entity + "." + action
}
