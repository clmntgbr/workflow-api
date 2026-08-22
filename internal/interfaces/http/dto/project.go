package dto

type CreateProjectRequest struct {
	Name string `json:"name" validate:"required,min=1,max=255"`
}

type UpdateProjectRequest struct {
	Name string `json:"name" validate:"required,min=1,max=255"`
}
