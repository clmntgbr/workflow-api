package invoice

import (
	"context"
	"encoding/json"

	"go-api/internal/application/mail"
	"go-api/internal/application/messaging"
	domaininvoice "go-api/internal/domain/invoice"
	domainuser "go-api/internal/domain/user"

	"github.com/google/uuid"
)

const PaymentMailHandlerName = "InvoicePaymentMailHandler"

type PaymentMailHandler struct {
	users domainuser.UserReadRepository
	mail  *mail.BillingMailService
}

func NewPaymentMailHandler(users domainuser.UserReadRepository, mailService *mail.BillingMailService) *PaymentMailHandler {
	return &PaymentMailHandler{users: users, mail: mailService}
}

func (h *PaymentMailHandler) OnSucceeded(ctx context.Context, payload []byte) error {
	var evt domaininvoice.InvoicePaymentSucceeded
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	to, err := h.recipients(ctx, evt.UserID)
	if err != nil {
		return messaging.Retryable(err)
	}
	if err := h.mail.NotifyPaymentSucceeded(ctx, evt, to); err != nil {
		return messaging.Retryable(err)
	}
	return nil
}

func (h *PaymentMailHandler) OnFailed(ctx context.Context, payload []byte) error {
	var evt domaininvoice.InvoicePaymentFailed
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	to, err := h.recipients(ctx, evt.UserID)
	if err != nil {
		return messaging.Retryable(err)
	}
	if err := h.mail.NotifyPaymentFailed(ctx, evt, to); err != nil {
		return messaging.Retryable(err)
	}
	return nil
}

func (h *PaymentMailHandler) recipients(ctx context.Context, userIDRaw string) ([]string, error) {
	userID, err := uuid.Parse(userIDRaw)
	if err != nil {
		return nil, err
	}
	user, err := h.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}
	return mail.RecipientsFromUser(user.Email, user.Banned), nil
}
