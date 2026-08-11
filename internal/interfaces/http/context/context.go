package context

import (
	"errors"

	domainuser "go-api/internal/domain/user"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

const UserKey = "user"

func GetUser(c fiber.Ctx) (*domainuser.User, error) {
	user, ok := c.Locals(UserKey).(*domainuser.User)
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func SetUser(c fiber.Ctx, user domainuser.User) {
	c.Locals(UserKey, &user)
}

func GetActiveOrganizationID(c fiber.Ctx) (uuid.UUID, error) {
	user, err := GetUser(c)
	if err != nil {
		return uuid.Nil, err
	}
	if user.ActiveOrganizationID == nil || *user.ActiveOrganizationID == uuid.Nil {
		return uuid.Nil, errors.New("active organization is required")
	}
	return *user.ActiveOrganizationID, nil
}
