package dto

type CreateOrganizationRequest struct {
	Name string `json:"name" validate:"required,min=1,max=255"`
}

type UpdateOrganizationRequest struct {
	Name string `json:"name" validate:"required,min=1,max=255"`
}

type AddOrganizationMemberRequest struct {
	UserID string `json:"userId" validate:"required,uuid"`
}
