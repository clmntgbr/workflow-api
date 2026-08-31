package handler

import (
	httpctx "go-api/internal/interfaces/http/context"

	"github.com/gofiber/fiber/v3"
)

type RealtimeHandler struct {
	connectionCreator realtimeConnectionCreator
}

func NewRealtimeHandler(connectionCreator realtimeConnectionCreator) *RealtimeHandler {
	return &RealtimeHandler{connectionCreator: connectionCreator}
}

func (h *RealtimeHandler) GetConnection(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	info, err := h.connectionCreator.CreateConnectionInfo(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to create realtime connection",
		})
	}

	return c.Status(fiber.StatusOK).JSON(info)
}
