package context

import (
	"errors"

	domainuser "go-api/internal/domain/user"

	"github.com/gofiber/fiber/v3"
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
