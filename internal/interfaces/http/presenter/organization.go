package presenter

import (
	"time"

	domainorganization "go-api/internal/domain/organization"

	"github.com/google/uuid"
)

type OrganizationDetailResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	IsActive  bool      `json:"isActive"`
	MemberIDs []string  `json:"memberIds"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func NewOrganizationDetailResponseFromView(
	view domainorganization.OrganizationView,
	activeOrganizationID *uuid.UUID,
) OrganizationDetailResponse {
	return OrganizationDetailResponse{
		ID:        view.ID.String(),
		Name:      view.Name,
		IsActive:  isActiveOrganization(view.ID, activeOrganizationID),
		MemberIDs: uuidStrings(view.MemberIDs),
		CreatedAt: view.CreatedAt,
		UpdatedAt: view.UpdatedAt,
	}
}

func NewOrganizationDetailResponseFromEntity(
	org domainorganization.Organization,
	activeOrganizationID *uuid.UUID,
) OrganizationDetailResponse {
	return OrganizationDetailResponse{
		ID:        org.ID.String(),
		Name:      org.Name,
		IsActive:  isActiveOrganization(org.ID, activeOrganizationID),
		MemberIDs: uuidStrings(org.MemberIDs),
		CreatedAt: org.CreatedAt,
		UpdatedAt: org.UpdatedAt,
	}
}

func NewOrganizationListResponseFromViews(
	views []domainorganization.OrganizationView,
	activeOrganizationID *uuid.UUID,
) []OrganizationDetailResponse {
	items := make([]OrganizationDetailResponse, 0, len(views))
	for _, view := range views {
		items = append(items, NewOrganizationDetailResponseFromView(view, activeOrganizationID))
	}
	return items
}

func isActiveOrganization(organizationID uuid.UUID, activeOrganizationID *uuid.UUID) bool {
	return activeOrganizationID != nil && *activeOrganizationID == organizationID
}

func uuidStrings(ids []uuid.UUID) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, id.String())
	}
	return out
}
