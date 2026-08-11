package handler

import (
	"strings"

	orgcmd "go-api/internal/application/command/organization"
	queryorganization "go-api/internal/application/query/organization"
	"go-api/internal/interfaces/http/dto"
	"go-api/internal/interfaces/http/presenter"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type OrganizationHandler struct {
	createHandler       *orgcmd.CreateOrganizationHandler
	updateHandler       *orgcmd.UpdateOrganizationHandler
	deleteHandler       *orgcmd.DeleteOrganizationHandler
	addMemberHandler    *orgcmd.AddOrganizationMemberHandler
	removeMemberHandler *orgcmd.RemoveOrganizationMemberHandler
	getByIDHandler      *queryorganization.GetOrganizationByIDHandler
}

func NewOrganizationHandler(
	createHandler *orgcmd.CreateOrganizationHandler,
	updateHandler *orgcmd.UpdateOrganizationHandler,
	deleteHandler *orgcmd.DeleteOrganizationHandler,
	addMemberHandler *orgcmd.AddOrganizationMemberHandler,
	removeMemberHandler *orgcmd.RemoveOrganizationMemberHandler,
	getByIDHandler *queryorganization.GetOrganizationByIDHandler,
) *OrganizationHandler {
	return &OrganizationHandler{
		createHandler:       createHandler,
		updateHandler:       updateHandler,
		deleteHandler:       deleteHandler,
		addMemberHandler:    addMemberHandler,
		removeMemberHandler: removeMemberHandler,
		getByIDHandler:      getByIDHandler,
	}
}

func (h *OrganizationHandler) Create(c fiber.Ctx) error {
	var req dto.CreateOrganizationRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request body"})
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "name is required"})
	}

	org, err := h.createHandler.Handle(c.Context(), orgcmd.CreateOrganizationCommand{Name: req.Name})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to create organization"})
	}

	return c.Status(fiber.StatusCreated).JSON(presenter.NewOrganizationDetailResponseFromEntity(*org))
}

func (h *OrganizationHandler) GetByID(c fiber.Ctx) error {
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

	return c.Status(fiber.StatusOK).JSON(presenter.NewOrganizationDetailResponseFromView(*view))
}

func (h *OrganizationHandler) Update(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid organization id"})
	}

	var req dto.UpdateOrganizationRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request body"})
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "name is required"})
	}
	if req.IsActive == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "isActive is required"})
	}

	err = h.updateHandler.Handle(c.Context(), orgcmd.UpdateOrganizationCommand{
		ID:       id,
		Name:     req.Name,
		IsActive: *req.IsActive,
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

	return c.Status(fiber.StatusOK).JSON(presenter.NewOrganizationDetailResponseFromView(*view))
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
