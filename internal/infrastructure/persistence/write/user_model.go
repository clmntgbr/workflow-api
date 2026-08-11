package write

import (
	"time"

	domainuser "go-api/internal/domain/user"

	"github.com/google/uuid"
)

// UserModel is the persistence mapping for table users.
// Schema is owned by SQL migrations — do not encode DDL in tags.
type UserModel struct {
	ID        uuid.UUID `gorm:"column:id;primaryKey"`
	ClerkID   string    `gorm:"column:clerk_id"`
	FirstName string    `gorm:"column:first_name"`
	LastName  string    `gorm:"column:last_name"`
	Banned    bool      `gorm:"column:banned"`
	Email     string    `gorm:"column:email"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (UserModel) TableName() string {
	return "users"
}

func userModelFromDomain(u *domainuser.User) *UserModel {
	return &UserModel{
		ID:        u.ID,
		ClerkID:   u.ClerkID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Banned:    u.Banned,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func userDomainFromModel(m *UserModel) *domainuser.User {
	return &domainuser.User{
		ID:        m.ID,
		ClerkID:   m.ClerkID,
		FirstName: m.FirstName,
		LastName:  m.LastName,
		Banned:    m.Banned,
		Email:     m.Email,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
