package presenter

import (
	"time"

	domainuser "go-api/internal/domain/user"

	"github.com/google/uuid"
)

type UserDetailResponse struct {
	ID                   string    `json:"id"`
	ClerkID              string    `json:"clerkId"`
	FirstName            string    `json:"firstName"`
	LastName             string    `json:"lastName"`
	Email                string    `json:"email"`
	ActiveProjectID *string   `json:"activeProjectId"`
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`
}

func optionalUUIDString(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	s := id.String()
	return &s
}

func NewUserDetailResponseFromView(view domainuser.UserView) UserDetailResponse {
	return UserDetailResponse{
		ID:                   view.ID.String(),
		ClerkID:              view.ClerkID,
		FirstName:            view.FirstName,
		LastName:             view.LastName,
		Email:                view.Email,
		ActiveProjectID: optionalUUIDString(view.ActiveProjectID),
		CreatedAt:            view.CreatedAt,
		UpdatedAt:            view.UpdatedAt,
	}
}

func NewUserDetailResponseFromEntity(user domainuser.User) UserDetailResponse {
	return UserDetailResponse{
		ID:                   user.ID.String(),
		ClerkID:              user.ClerkID,
		FirstName:            user.FirstName,
		LastName:             user.LastName,
		Email:                user.Email,
		ActiveProjectID: optionalUUIDString(user.ActiveProjectID),
		CreatedAt:            user.CreatedAt,
		UpdatedAt:            user.UpdatedAt,
	}
}
