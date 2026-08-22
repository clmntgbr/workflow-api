package invoice

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	domaininvoice "go-api/internal/domain/invoice"
)

type InvoiceCreatedHandler struct{}

func NewInvoiceCreatedHandler() *InvoiceCreatedHandler {
	return &InvoiceCreatedHandler{}
}

func (h *InvoiceCreatedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domaininvoice.InvoiceCreated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	log.Printf(
		"event handled %s eventId=%s invoiceId=%s userId=%s stripeInvoiceId=%s status=%s total=%d",
		domaininvoice.EventTypeInvoiceCreated,
		evt.ID,
		evt.InvoiceID,
		evt.UserID,
		evt.StripeInvoiceID,
		evt.Status,
		evt.Total,
	)
	return nil
}
