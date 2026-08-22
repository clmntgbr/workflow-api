package middleware

import (
	"encoding/json"
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
)

type BillingWebhookMiddleware struct {
	secret string
}

func NewBillingWebhookMiddleware(secret string) *BillingWebhookMiddleware {
	return &BillingWebhookMiddleware{
		secret: secret,
	}
}

func (m *BillingWebhookMiddleware) Protected() fiber.Handler {
	return func(c fiber.Ctx) error {
		log.Printf("stripe webhook: received %s %s", c.Method(), c.Path())

		if m.secret == "" {
			log.Printf("stripe webhook: secret is not configured")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "stripe webhook secret is not configured",
			})
		}

		payload := c.Body()
		signature := c.Get("Stripe-Signature")
		if signature == "" {
			log.Printf("stripe webhook: missing Stripe-Signature header")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing stripe signature",
			})
		}

		if err := webhook.ValidatePayload(payload, signature, m.secret); err != nil {
			log.Printf("stripe webhook: invalid signature: %v", err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid signature",
			})
		}

		var event stripe.Event
		if err := json.Unmarshal(payload, &event); err != nil {
			log.Printf("stripe webhook: invalid payload: %v", err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid payload",
			})
		}

		log.Printf("stripe webhook: accepted event id=%s type=%s", event.ID, event.Type)
		c.Locals("payload", event)

		return c.Next()
	}
}
