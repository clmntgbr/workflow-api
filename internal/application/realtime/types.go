package realtime

const (
	ActionCreated                   = "created"
	ActionUpdated                   = "updated"
	ActionDeleted                   = "deleted"
	ActionActiveOrganizationChanged = "active_organization_changed"
	ActionMemberAdded               = "member_added"
	ActionMemberRemoved             = "member_removed"

	EntityUser         = "user"
	EntityOrganization = "organization"
	EntityWorkflow     = "workflow"
	EntityEndpoint     = "endpoint"
	EntityStep         = "step"
	EntityConnection   = "connection"
)

func EventType(entity, action string) string {
	return entity + "." + action
}
