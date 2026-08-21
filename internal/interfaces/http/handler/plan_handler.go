package handler

import (
	"errors"

	queryplan "go-api/internal/application/query/plan"
	domainplan "go-api/internal/domain/plan"
	"go-api/internal/interfaces/http/presenter"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type PlanHandler struct {
	listActive *queryplan.ListActivePlansHandler
	getByID    *queryplan.GetPlanByIDHandler
}

func NewPlanHandler(
	listActive *queryplan.ListActivePlansHandler,
	getByID *queryplan.GetPlanByIDHandler,
) *PlanHandler {
	return &PlanHandler{
		listActive: listActive,
		getByID:    getByID,
	}
}

func (h *PlanHandler) List(c fiber.Ctx) error {
	views, err := h.listActive.Handle(c.Context(), queryplan.ListActivePlansQuery{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to list plans"})
	}
	return c.Status(fiber.StatusOK).JSON(presenter.NewPlanListResponseFromViews(views))
}

func (h *PlanHandler) GetByID(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid plan id"})
	}
	view, err := h.getByID.Handle(c.Context(), queryplan.GetPlanByIDQuery{ID: id})
	if err != nil {
		if errors.Is(err, domainplan.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Plan not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to get plan"})
	}
	return c.Status(fiber.StatusOK).JSON(presenter.NewPlanResponseFromView(*view))
}
