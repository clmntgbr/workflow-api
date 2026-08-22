package dto

type SetActiveProjectRequest struct {
	ProjectID string `json:"projectId" validate:"required,uuid"`
}
