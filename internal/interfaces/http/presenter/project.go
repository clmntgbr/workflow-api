package presenter

import (
	"time"

	domainproject "go-api/internal/domain/project"

	"github.com/google/uuid"
)

type ProjectDetailResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	IsActive  bool      `json:"isActive"`
	MemberIDs []string  `json:"memberIds"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func NewProjectDetailResponseFromView(
	view domainproject.ProjectView,
	activeProjectID *uuid.UUID,
) ProjectDetailResponse {
	return ProjectDetailResponse{
		ID:        view.ID.String(),
		Name:      view.Name,
		IsActive:  isActiveProject(view.ID, activeProjectID),
		MemberIDs: uuidStrings(view.MemberIDs),
		CreatedAt: view.CreatedAt,
		UpdatedAt: view.UpdatedAt,
	}
}

func NewProjectDetailResponseFromEntity(
	org domainproject.Project,
	activeProjectID *uuid.UUID,
) ProjectDetailResponse {
	return ProjectDetailResponse{
		ID:        org.ID.String(),
		Name:      org.Name,
		IsActive:  isActiveProject(org.ID, activeProjectID),
		MemberIDs: uuidStrings(org.MemberIDs),
		CreatedAt: org.CreatedAt,
		UpdatedAt: org.UpdatedAt,
	}
}

func NewProjectListResponseFromViews(
	views []domainproject.ProjectView,
	activeProjectID *uuid.UUID,
) []ProjectDetailResponse {
	items := make([]ProjectDetailResponse, 0, len(views))
	for _, view := range views {
		items = append(items, NewProjectDetailResponseFromView(view, activeProjectID))
	}
	return items
}

func isActiveProject(projectID uuid.UUID, activeProjectID *uuid.UUID) bool {
	return activeProjectID != nil && *activeProjectID == projectID
}

func uuidStrings(ids []uuid.UUID) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, id.String())
	}
	return out
}
