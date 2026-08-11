package handler

import (
	"strings"

	orgcmd "go-api/internal/application/command/organization"
	usercmd "go-api/internal/application/command/user"
	queryorganization "go-api/internal/application/query/organization"
	httpctx "go-api/internal/interfaces/http/context"
	"go-api/internal/interfaces/http/dto"
	"go-api/internal/interfaces/http/presenter"
	"go-api/internal/interfaces/http/validation"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type OrganizationHandler struct {
	createHandler                *orgcmd.CreateOrganizationHandler
	updateHandler                *orgcmd.UpdateOrganizationHandler
	deleteHandler                *orgcmd.DeleteOrganizationHandler
	addMemberHandler             *orgcmd.AddOrganizationMemberHandler
	removeMemberHandler          *orgcmd.RemoveOrganizationMemberHandler
	getByIDHandler               *queryorganization.GetOrganizationByIDHandler
	listByUserHandler            *queryorganization.ListOrganizationsByUserHandler
	setActiveOrganizationHandler *usercmd.SetActiveOrganizationHandler
}

func NewOrganizationHandler(
	createHandler *orgcmd.CreateOrganizationHandler,
	updateHandler *orgcmd.UpdateOrganizationHandler,
	deleteHandler *orgcmd.DeleteOrganizationHandler,
	addMemberHandler *orgcmd.AddOrganizationMemberHandler,
	removeMemberHandler *orgcmd.RemoveOrganizationMemberHandler,
	getByIDHandler *queryorganization.GetOrganizationByIDHandler,
	listByUserHandler *queryorganization.ListOrganizationsByUserHandler,
	setActiveOrganizationHandler *usercmd.SetActiveOrganizationHandler,
) *OrganizationHandler {
	return &OrganizationHandler{
		createHandler:                createHandler,
		updateHandler:                updateHandler,
		deleteHandler:                deleteHandler,
		addMemberHandler:             addMemberHandler,
		removeMemberHandler:          removeMemberHandler,
		getByIDHandler:               getByIDHandler,
		listByUserHandler:            listByUserHandler,
		setActiveOrganizationHandler: setActiveOrganizationHandler,
	}
}

func (h *OrganizationHandler) List(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	views, err := h.listByUserHandler.Handle(c.Context(), queryorganization.ListOrganizationsByUserQuery{
		UserID: user.ID,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to list organizations"})
	}

	return c.Status(fiber.StatusOK).JSON(
		presenter.NewOrganizationListResponseFromViews(views, user.ActiveOrganizationID),
	)
}

func (h *OrganizationHandler) Create(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	var req dto.CreateOrganizationRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request body"})
	}
	req.Name = strings.TrimSpace(req.Name)
	if err := validation.Struct(c, &req); err != nil {
		return err
	}

	org, err := h.createHandler.Handle(c.Context(), orgcmd.CreateOrganizationCommand{
		Name:          req.Name,
		CreatorUserID: user.ID,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to create organization"})
	}

	activeID := org.ID
	return c.Status(fiber.StatusCreated).JSON(
		presenter.NewOrganizationDetailResponseFromEntity(*org, &activeID),
	)
}

func (h *OrganizationHandler) GetByID(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid organization id"})
	}

	view, err := h.getByIDHandler.Handle(c.Context(), queryorganization.GetOrganizationByIDQuery{ID: id})
	if err != nil {
		if err.Error() == "organization not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Organization not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to get organization"})
	}

	return c.Status(fiber.StatusOK).JSON(
		presenter.NewOrganizationDetailResponseFromView(*view, user.ActiveOrganizationID),
	)
}

func (h *OrganizationHandler) Update(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid organization id"})
	}

	var req dto.UpdateOrganizationRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request body"})
	}
	req.Name = strings.TrimSpace(req.Name)
	if err := validation.Struct(c, &req); err != nil {
		return err
	}

	err = h.updateHandler.Handle(c.Context(), orgcmd.UpdateOrganizationCommand{
		ID:   id,
		Name: req.Name,
	})
	if err != nil {
		if err.Error() == "organization not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Organization not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to update organization"})
	}

	view, err := h.getByIDHandler.Handle(c.Context(), queryorganization.GetOrganizationByIDQuery{ID: id})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to get organization"})
	}

	return c.Status(fiber.StatusOK).JSON(
		presenter.NewOrganizationDetailResponseFromView(*view, user.ActiveOrganizationID),
	)
}

func (h *OrganizationHandler) Activate(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid organization id"})
	}

	err = h.setActiveOrganizationHandler.Handle(c.Context(), usercmd.SetActiveOrganizationCommand{
		UserID:         user.ID,
		OrganizationID: id,
	})
	if err != nil {
		switch err.Error() {
		case "organization not found":
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Organization not found"})
		case "user is not a member of the organization":
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "User is not a member of the organization"})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to activate organization"})
		}
	}

	view, err := h.getByIDHandler.Handle(c.Context(), queryorganization.GetOrganizationByIDQuery{ID: id})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to get organization"})
	}

	activeID := id
	return c.Status(fiber.StatusOK).JSON(
		presenter.NewOrganizationDetailResponseFromView(*view, &activeID),
	)
}

func (h *OrganizationHandler) Delete(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid organization id"})
	}

	if err := h.deleteHandler.Handle(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to delete organization"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *OrganizationHandler) AddMember(c fiber.Ctx) error {
	orgID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid organization id"})
	}

	var req dto.AddOrganizationMemberRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request body"})
	}
	if err := validation.Struct(c, &req); err != nil {
		return err
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid user id"})
	}

	err = h.addMemberHandler.Handle(c.Context(), orgcmd.AddOrganizationMemberCommand{
		OrganizationID: orgID,
		UserID:         userID,
	})
	if err != nil {
		if err.Error() == "organization not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Organization not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to add member"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *OrganizationHandler) RemoveMember(c fiber.Ctx) error {
	orgID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid organization id"})
	}
	userID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid user id"})
	}

	err = h.removeMemberHandler.Handle(c.Context(), orgcmd.RemoveOrganizationMemberCommand{
		OrganizationID: orgID,
		UserID:         userID,
	})
	if err != nil {
		if err.Error() == "organization not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Organization not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to remove member"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
