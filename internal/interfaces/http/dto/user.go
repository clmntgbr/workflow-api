package dto

type SetActiveOrganizationRequest struct {
	OrganizationID string `json:"organizationId" validate:"required,uuid"`
}
