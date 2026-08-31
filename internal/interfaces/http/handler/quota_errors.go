package handler

import (
	"errors"

	cmdquota "go-api/internal/application/command/quota"
	querysubscription "go-api/internal/application/query/subscription"

	"github.com/gofiber/fiber/v3"
)

func respondQuotaError(c fiber.Ctx, err error) (bool, error) {
	if err == nil {
		return false, nil
	}

	switch {
	case errors.Is(err, querysubscription.ErrSubscriptionNotFound):
		return true, c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "No active subscription found",
		})
	case errors.Is(err, querysubscription.ErrActiveProjectRequired):
		return true, c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Active project is required",
		})
	case errors.Is(err, cmdquota.ErrWorkflowQuotaExceeded),
		errors.Is(err, cmdquota.ErrEndpointQuotaExceeded),
		errors.Is(err, cmdquota.ErrStepQuotaExceeded),
		errors.Is(err, cmdquota.ErrVariableQuotaExceeded),
		errors.Is(err, cmdquota.ErrAssertionQuotaExceeded),
		errors.Is(err, cmdquota.ErrWorkflowRunQuotaExceeded),
		errors.Is(err, cmdquota.ErrConcurrentRunQuotaExceeded),
		errors.Is(err, cmdquota.ErrProjectQuotaExceeded),
		errors.Is(err, cmdquota.ErrScheduleIntervalQuotaExceeded):
		return true, c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": err.Error(),
		})
	default:
		return false, nil
	}
}
