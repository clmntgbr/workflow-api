package handler

import (
	"errors"

	cmdsubscription "go-api/internal/application/command/subscription"
	querysubscription "go-api/internal/application/query/subscription"
	httpctx "go-api/internal/interfaces/http/context"
	"go-api/internal/interfaces/http/dto"
	"go-api/internal/interfaces/http/presenter"
	"go-api/internal/interfaces/http/validation"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type SubscriptionHandler struct {
	getCurrentSubscriptionHandler subscriptionGetCurrentHandler
	getQuotaUsageHandler          subscriptionGetQuotaHandler
	previewPlanChangeHandler      subscriptionPreviewPlanChangeHandler
	createSubscriptionHandler     subscriptionCreateHandler
	createBillingPortalHandler    subscriptionCreateBillingPortalHandler
}

func NewSubscriptionHandler(
	getCurrentSubscriptionHandler subscriptionGetCurrentHandler,
	getQuotaUsageHandler subscriptionGetQuotaHandler,
	previewPlanChangeHandler subscriptionPreviewPlanChangeHandler,
	createSubscriptionHandler subscriptionCreateHandler,
	createBillingPortalHandler subscriptionCreateBillingPortalHandler,
) *SubscriptionHandler {
	return &SubscriptionHandler{
		getCurrentSubscriptionHandler: getCurrentSubscriptionHandler,
		getQuotaUsageHandler:          getQuotaUsageHandler,
		previewPlanChangeHandler:      previewPlanChangeHandler,
		createSubscriptionHandler:     createSubscriptionHandler,
		createBillingPortalHandler:    createBillingPortalHandler,
	}
}

func (h *SubscriptionHandler) GetSubscription(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	view, err := h.getCurrentSubscriptionHandler.Handle(c.Context(), querysubscription.GetCurrentSubscriptionQuery{
		UserID: user.ID,
	})
	if err != nil {
		if errors.Is(err, querysubscription.ErrSubscriptionNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Subscription not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to get subscription",
		})
	}

	return c.Status(fiber.StatusOK).JSON(presenter.NewSubscriptionResponse(view))
}

func (h *SubscriptionHandler) GetQuota(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	orgID, err := httpctx.GetActiveProjectID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Active project is required",
		})
	}

	usage, err := h.getQuotaUsageHandler.Handle(c.Context(), querysubscription.GetQuotaUsageQuery{
		UserID:         user.ID,
		ProjectID: orgID,
	})
	if err != nil {
		switch {
		case errors.Is(err, querysubscription.ErrSubscriptionNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Subscription not found",
			})
		case errors.Is(err, querysubscription.ErrActiveProjectRequired):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Active project is required",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Failed to get quota usage",
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(presenter.NewQuotaUsageResponse(usage))
}

func (h *SubscriptionHandler) PreviewSubscription(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	var request dto.PreviewSubscriptionRequest
	if err := validation.BindBody(c, &request); err != nil {
		return err
	}

	planID, err := uuid.Parse(request.PlanID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid planId",
		})
	}

	preview, err := h.previewPlanChangeHandler.Handle(c.Context(), querysubscription.PreviewPlanChangeQuery{
		UserID: user.ID,
		PlanID: planID,
	})
	if err != nil {
		switch {
		case errors.Is(err, querysubscription.ErrPlanNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Plan not found",
			})
		case errors.Is(err, querysubscription.ErrPlanInactive),
			errors.Is(err, querysubscription.ErrFreePlanCheckout),
			errors.Is(err, querysubscription.ErrMissingStripePrice),
			errors.Is(err, querysubscription.ErrAlreadyOnPlan):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": err.Error(),
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Failed to preview plan change",
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(presenter.NewPlanChangePreviewResponse(preview))
}

func (h *SubscriptionHandler) CreateSubscription(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	var request dto.CreateSubscriptionRequest
	if err := validation.BindBody(c, &request); err != nil {
		return err
	}

	planID, err := uuid.Parse(request.PlanID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid planId",
		})
	}

	result, err := h.createSubscriptionHandler.Handle(c.Context(), cmdsubscription.CreateSubscriptionCommand{
		UserID:        user.ID,
		PlanID:        planID,
		ProrationDate: request.ProrationDate,
	})
	if err != nil {
		switch {
		case errors.Is(err, querysubscription.ErrPlanNotFound),
			errors.Is(err, querysubscription.ErrSubscriptionNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": err.Error(),
			})
		case errors.Is(err, querysubscription.ErrPlanInactive),
			errors.Is(err, querysubscription.ErrFreePlanCheckout),
			errors.Is(err, querysubscription.ErrMissingStripePrice),
			errors.Is(err, querysubscription.ErrAlreadyOnPlan):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": err.Error(),
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Internal server error",
				"errors":  err.Error(),
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(presenter.NewChangeSubscriptionResponse(result))
}

func (h *SubscriptionHandler) CreateBillingPortal(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	url, err := h.createBillingPortalHandler.Handle(c.Context(), cmdsubscription.CreateBillingPortalCommand{
		UserID: user.ID,
	})
	if err != nil {
		switch {
		case errors.Is(err, querysubscription.ErrSubscriptionNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Subscription not found",
			})
		case errors.Is(err, cmdsubscription.ErrMissingStripeCustomer):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "User has no Stripe customer",
			})
		case errors.Is(err, cmdsubscription.ErrFreePlanBillingPortal):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": err.Error(),
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Internal server error",
				"errors":  err.Error(),
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(presenter.NewCheckoutSessionResponse(url))
}
