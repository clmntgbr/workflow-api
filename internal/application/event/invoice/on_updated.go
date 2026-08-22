package invoice

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	domaininvoice "go-api/internal/domain/invoice"
)

type InvoiceUpdatedHandler struct{}

func NewInvoiceUpdatedHandler() *InvoiceUpdatedHandler {
	return &InvoiceUpdatedHandler{}
}

func (h *InvoiceUpdatedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domaininvoice.InvoiceUpdated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	log.Printf(
		"event handled %s eventId=%s invoiceId=%s userId=%s stripeInvoiceId=%s status=%s total=%d",
		domaininvoice.EventTypeInvoiceUpdated,
		evt.ID,
		evt.InvoiceID,
		evt.UserID,
		evt.StripeInvoiceID,
		evt.Status,
		evt.Total,
	)
	return nil
}
