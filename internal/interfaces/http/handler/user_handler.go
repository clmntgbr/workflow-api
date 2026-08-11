package handler

import (
	usercmd "go-api/internal/application/command/user"
	queryuser "go-api/internal/application/query/user"
	httpctx "go-api/internal/interfaces/http/context"
	"go-api/internal/interfaces/http/dto"
	"go-api/internal/interfaces/http/presenter"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type UserHandler struct {
	getUserByIDHandler           *queryuser.GetUserByIDHandler
	setActiveOrganizationHandler *usercmd.SetActiveOrganizationHandler
}

func NewUserHandler(
	getUserByIDHandler *queryuser.GetUserByIDHandler,
	setActiveOrganizationHandler *usercmd.SetActiveOrganizationHandler,
) *UserHandler {
	return &UserHandler{
		getUserByIDHandler:           getUserByIDHandler,
		setActiveOrganizationHandler: setActiveOrganizationHandler,
	}
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

func (h *UserHandler) SetActiveOrganization(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	var req dto.SetActiveOrganizationRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request body"})
	}

	orgID, err := uuid.Parse(req.OrganizationID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid organization id"})
	}

	err = h.setActiveOrganizationHandler.Handle(c.Context(), usercmd.SetActiveOrganizationCommand{
		UserID:         user.ID,
		OrganizationID: orgID,
	})
	if err != nil {
		switch err.Error() {
		case "organization not found":
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Organization not found"})
		case "user is not a member of the organization":
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "User is not a member of the organization"})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to set active organization"})
		}
	}

	view, err := h.getUserByIDHandler.Handle(c.Context(), queryuser.GetUserByIDQuery{ID: user.ID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to get user"})
	}

	return c.Status(fiber.StatusOK).JSON(presenter.NewUserDetailResponseFromView(*view))
}
