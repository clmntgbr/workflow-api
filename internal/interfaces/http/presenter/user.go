package presenter

import (
	"time"

	domainuser "go-api/internal/domain/user"
)

type UserDetailResponse struct {
	ID        string    `json:"id"`
	ClerkID   string    `json:"clerkId"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func NewUserDetailResponseFromView(view domainuser.UserView) UserDetailResponse {
	return UserDetailResponse{
		ID:        view.ID.String(),
		ClerkID:   view.ClerkID,
		FirstName: view.FirstName,
		LastName:  view.LastName,
		Email:     view.Email,
		CreatedAt: view.CreatedAt,
		UpdatedAt: view.UpdatedAt,
	}
}

func NewUserDetailResponseFromEntity(user domainuser.User) UserDetailResponse {
	return UserDetailResponse{
		ID:        user.ID.String(),
		ClerkID:   user.ClerkID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
