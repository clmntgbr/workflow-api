package dto

type CreateOrganizationRequest struct {
	Name string `json:"name"`
}

type UpdateOrganizationRequest struct {
	Name     string `json:"name"`
	IsActive *bool  `json:"isActive"`
}

type AddOrganizationMemberRequest struct {
	UserID string `json:"userId"`
}
