package subscription

import (
	"context"
	"encoding/json"

	"go-api/internal/application/mail"
	"go-api/internal/application/messaging"
	domainplan "go-api/internal/domain/plan"
	domainsubscription "go-api/internal/domain/subscription"
	domainuser "go-api/internal/domain/user"

	"github.com/google/uuid"
)

const BillingMailHandlerName = "SubscriptionBillingMailHandler"

type BillingMailHandler struct {
	users domainuser.UserReadRepository
	plans domainplan.PlanReadRepository
	mail  *mail.BillingMailService
}

func NewBillingMailHandler(
	users domainuser.UserReadRepository,
	plans domainplan.PlanReadRepository,
	mailService *mail.BillingMailService,
) *BillingMailHandler {
	return &BillingMailHandler{users: users, plans: plans, mail: mailService}
}

func (h *BillingMailHandler) OnPlanChanged(ctx context.Context, payload []byte) error {
	var evt domainsubscription.SubscriptionPlanChanged
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	to, err := h.recipientsBySubscription(ctx, evt.SubscriptionID)
	if err != nil {
		return err
	}
	previousName, err := h.planName(ctx, evt.PreviousPlanID)
	if err != nil {
		return messaging.Retryable(err)
	}
	newName, err := h.planName(ctx, evt.PlanID)
	if err != nil {
		return messaging.Retryable(err)
	}
	if err := h.mail.NotifyPlanChanged(ctx, previousName, newName, to); err != nil {
		return messaging.Retryable(err)
	}
	return nil
}

func (h *BillingMailHandler) OnRenewalUpcoming(ctx context.Context, payload []byte) error {
	var evt domainsubscription.SubscriptionRenewalUpcoming
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	to, err := h.recipientsBySubscription(ctx, evt.SubscriptionID)
	if err != nil {
		return err
	}
	planName, err := h.planName(ctx, evt.PlanID)
	if err != nil {
		return messaging.Retryable(err)
	}
	if err := h.mail.NotifyRenewalUpcoming(ctx, evt, planName, to); err != nil {
		return messaging.Retryable(err)
	}
	return nil
}

func (h *BillingMailHandler) OnPaymentMethodExpiring(ctx context.Context, payload []byte) error {
	var evt domainsubscription.SubscriptionPaymentMethodExpiring
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	to, err := h.recipientsBySubscription(ctx, evt.SubscriptionID)
	if err != nil {
		return err
	}
	if err := h.mail.NotifyPaymentMethodExpiring(ctx, evt, to); err != nil {
		return messaging.Retryable(err)
	}
	return nil
}

func (h *BillingMailHandler) recipientsBySubscription(ctx context.Context, subscriptionIDRaw string) ([]string, error) {
	subscriptionID, err := uuid.Parse(subscriptionIDRaw)
	if err != nil {
		return nil, messaging.NonRetryable(err)
	}
	user, err := h.users.FindBySubscriptionID(ctx, subscriptionID)
	if err != nil {
		return nil, messaging.Retryable(err)
	}
	if user == nil {
		return nil, nil
	}
	return mail.RecipientsFromUser(user.Email, user.Banned), nil
}

func (h *BillingMailHandler) planName(ctx context.Context, planIDRaw string) (string, error) {
	planID, err := uuid.Parse(planIDRaw)
	if err != nil {
		return "", err
	}
	plan, err := h.plans.FindByID(ctx, planID)
	if err != nil {
		return "", err
	}
	if plan == nil {
		return "", nil
	}
	return plan.Name, nil
}
