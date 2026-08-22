package handler

import (
	usercmd "go-api/internal/application/command/user"
	queryuser "go-api/internal/application/query/user"
	httpctx "go-api/internal/interfaces/http/context"
	"go-api/internal/interfaces/http/dto"
	"go-api/internal/interfaces/http/presenter"
	"go-api/internal/interfaces/http/validation"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type UserHandler struct {
	getUserByIDHandler           *queryuser.GetUserByIDHandler
	setActiveProjectHandler *usercmd.SetActiveProjectHandler
}

func NewUserHandler(
	getUserByIDHandler *queryuser.GetUserByIDHandler,
	setActiveProjectHandler *usercmd.SetActiveProjectHandler,
) *UserHandler {
	return &UserHandler{
		getUserByIDHandler:           getUserByIDHandler,
		setActiveProjectHandler: setActiveProjectHandler,
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

func (h *UserHandler) SetActiveProject(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	var req dto.SetActiveProjectRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request body"})
	}
	if err := validation.Struct(c, &req); err != nil {
		return err
	}

	orgID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid project id"})
	}

	err = h.setActiveProjectHandler.Handle(c.Context(), usercmd.SetActiveProjectCommand{
		UserID:         user.ID,
		ProjectID: orgID,
	})
	if err != nil {
		switch err.Error() {
		case "project not found":
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Project not found"})
		case "user is not a member of the project":
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "User is not a member of the project"})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to set active project"})
		}
	}

	view, err := h.getUserByIDHandler.Handle(c.Context(), queryuser.GetUserByIDQuery{ID: user.ID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to get user"})
	}

	return c.Status(fiber.StatusOK).JSON(presenter.NewUserDetailResponseFromView(*view))
}
