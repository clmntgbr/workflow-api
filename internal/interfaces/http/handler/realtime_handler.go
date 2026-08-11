package handler

import (
	"go-api/internal/infrastructure/centrifugo"
	"go-api/internal/infrastructure/config"
	httpctx "go-api/internal/interfaces/http/context"

	"github.com/gofiber/fiber/v3"
)

type RealtimeHandler struct {
	env *config.Config
}

func NewRealtimeHandler(env *config.Config) *RealtimeHandler {
	return &RealtimeHandler{env: env}
}

func (h *RealtimeHandler) GetConnection(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	info, err := centrifugo.NewConnectionInfo(h.env, user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to create realtime connection",
		})
	}

	return c.Status(fiber.StatusOK).JSON(info)
}
