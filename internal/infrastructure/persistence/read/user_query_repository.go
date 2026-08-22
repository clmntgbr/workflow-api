package read

import (
	"context"
	"errors"
	"time"

	domainuser "go-api/internal/domain/user"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userRow struct {
	ID                   uuid.UUID
	ClerkID              string
	FirstName            string
	LastName             string
	Email                string
	Banned               bool
	ActiveProjectID *uuid.UUID
	SubscriptionID       *uuid.UUID
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

func (userRow) TableName() string { return "users" }

type userReadRepository struct {
	db *gorm.DB
}

func NewUserReadRepository(db *gorm.DB) domainuser.UserReadRepository {
	return &userReadRepository{db: db}
}

func (r *userReadRepository) FindByID(ctx context.Context, id uuid.UUID) (*domainuser.UserView, error) {
	var row userRow
	err := r.db.WithContext(ctx).
		Select("id", "clerk_id", "first_name", "last_name", "email", "banned", "active_project_id", "subscription_id", "created_at", "updated_at").
		First(&row, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toUserView(row), nil
}

func (r *userReadRepository) FindByClerkID(ctx context.Context, clerkID string) (*domainuser.UserView, error) {
	var row userRow
	err := r.db.WithContext(ctx).
		Select("id", "clerk_id", "first_name", "last_name", "email", "banned", "active_project_id", "subscription_id", "created_at", "updated_at").
		Where("clerk_id = ?", clerkID).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toUserView(row), nil
}

func toUserView(row userRow) *domainuser.UserView {
	return &domainuser.UserView{
		ID:                   row.ID,
		ClerkID:              row.ClerkID,
		FirstName:            row.FirstName,
		LastName:             row.LastName,
		Email:                row.Email,
		Banned:               row.Banned,
		ActiveProjectID: row.ActiveProjectID,
		SubscriptionID:       row.SubscriptionID,
		CreatedAt:            row.CreatedAt,
		UpdatedAt:            row.UpdatedAt,
	}
}
