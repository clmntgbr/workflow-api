package handler

import "github.com/gofiber/fiber/v3"

// RespondQuotaErrorForTest exposes respondQuotaError for isolated mapper tests.
func RespondQuotaErrorForTest(c fiber.Ctx, err error) (bool, error) {
	return respondQuotaError(c, err)
}
