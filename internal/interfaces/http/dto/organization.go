package dto

type CreateOrganizationRequest struct {
	Name string `json:"name"`
}

type UpdateOrganizationRequest struct {
	Name string `json:"name"`
}

type AddOrganizationMemberRequest struct {
	UserID string `json:"userId"`
}
