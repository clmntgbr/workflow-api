package quota

import (
	"context"
	"encoding/json"

	"go-api/internal/application/mail"
	"go-api/internal/application/messaging"
	domainquota "go-api/internal/domain/quota"
	domainuser "go-api/internal/domain/user"

	"github.com/google/uuid"
)

const MailHandlerName = "QuotaMailHandler"

type MailHandler struct {
	users domainuser.UserReadRepository
	mail  *mail.QuotaMailService
}

func NewMailHandler(users domainuser.UserReadRepository, mailService *mail.QuotaMailService) *MailHandler {
	return &MailHandler{users: users, mail: mailService}
}

func (h *MailHandler) OnThresholdReached(ctx context.Context, payload []byte) error {
	var evt domainquota.QuotaThresholdReached
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	to, err := h.recipients(ctx, evt.UserID)
	if err != nil {
		return err
	}
	if err := h.mail.NotifyThresholdReached(ctx, evt, to); err != nil {
		return messaging.Retryable(err)
	}
	return nil
}

func (h *MailHandler) OnExceeded(ctx context.Context, payload []byte) error {
	var evt domainquota.QuotaExceeded
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	to, err := h.recipients(ctx, evt.UserID)
	if err != nil {
		return err
	}
	if err := h.mail.NotifyExceeded(ctx, evt, to); err != nil {
		return messaging.Retryable(err)
	}
	return nil
}

func (h *MailHandler) recipients(ctx context.Context, userIDRaw string) ([]string, error) {
	userID, err := uuid.Parse(userIDRaw)
	if err != nil {
		return nil, messaging.NonRetryable(err)
	}
	user, err := h.users.FindByID(ctx, userID)
	if err != nil {
		return nil, messaging.Retryable(err)
	}
	if user == nil {
		return nil, nil
	}
	return mail.RecipientsFromUser(user.Email, user.Banned), nil
}
