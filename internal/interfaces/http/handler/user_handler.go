package handler

import (
	queryuser "go-api/internal/application/query/user"
	httpctx "go-api/internal/interfaces/http/context"
	"go-api/internal/interfaces/http/presenter"

	"github.com/gofiber/fiber/v3"
)

type UserHandler struct {
	getUserByIDHandler *queryuser.GetUserByIDHandler
}

func NewUserHandler(getUserByIDHandler *queryuser.GetUserByIDHandler) *UserHandler {
	return &UserHandler{getUserByIDHandler: getUserByIDHandler}
}

func (h *UserHandler) GetUser(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	view, err := h.getUserByIDHandler.Handle(c.Context(), queryuser.GetUserByIDQuery{ID: user.ID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to get user",
		})
	}

	return c.Status(fiber.StatusOK).JSON(presenter.NewUserDetailResponseFromView(*view))
}
