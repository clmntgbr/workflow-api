package handler

import (
	queryplan "go-api/internal/application/query/plan"
	"go-api/internal/interfaces/http/presenter"

	"github.com/gofiber/fiber/v3"
)

type PlanHandler struct {
	listActive *queryplan.ListActivePlansHandler
}

func NewPlanHandler(listActive *queryplan.ListActivePlansHandler) *PlanHandler {
	return &PlanHandler{listActive: listActive}
}

func (h *PlanHandler) List(c fiber.Ctx) error {
	views, err := h.listActive.Handle(c.Context(), queryplan.ListActivePlansQuery{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to list plans"})
	}
	return c.Status(fiber.StatusOK).JSON(presenter.NewPlanListResponseFromViews(views))
}
