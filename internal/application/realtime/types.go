package realtime

const (
	ActionCreated                   = "created"
	ActionUpdated                   = "updated"
	ActionDeleted                   = "deleted"
	ActionActiveOrganizationChanged = "active_organization_changed"

	EntityUser         = "user"
	EntityOrganization = "organization"
	EntityWorkflow     = "workflow"
)

func EventType(entity, action string) string {
	return entity + "." + action
}
