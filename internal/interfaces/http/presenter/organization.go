package presenter

import (
	"time"

	domainorganization "go-api/internal/domain/organization"
)

type OrganizationDetailResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	IsActive  bool      `json:"isActive"`
	MemberIDs []string  `json:"memberIds"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func NewOrganizationDetailResponseFromView(view domainorganization.OrganizationView) OrganizationDetailResponse {
	memberIDs := make([]string, 0, len(view.MemberIDs))
	for _, id := range view.MemberIDs {
		memberIDs = append(memberIDs, id.String())
	}
	return OrganizationDetailResponse{
		ID:        view.ID.String(),
		Name:      view.Name,
		IsActive:  view.IsActive,
		MemberIDs: memberIDs,
		CreatedAt: view.CreatedAt,
		UpdatedAt: view.UpdatedAt,
	}
}

func NewOrganizationDetailResponseFromEntity(org domainorganization.Organization) OrganizationDetailResponse {
	memberIDs := make([]string, 0, len(org.MemberIDs))
	for _, id := range org.MemberIDs {
		memberIDs = append(memberIDs, id.String())
	}
	return OrganizationDetailResponse{
		ID:        org.ID.String(),
		Name:      org.Name,
		IsActive:  org.IsActive,
		MemberIDs: memberIDs,
		CreatedAt: org.CreatedAt,
		UpdatedAt: org.UpdatedAt,
	}
}
