package plan

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrNotFound      = errors.New("plan not found")
	ErrDuplicateSlug = errors.New("plan slug already exists")
)

func validatePlanInput(
	name, slug, stripePriceID string,
	interval BillingInterval,
	currency Currency,
	quotaID uuid.UUID,
) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("name is required")
	}
	if strings.TrimSpace(slug) == "" {
		return errors.New("slug is required")
	}
	if strings.TrimSpace(stripePriceID) == "" {
		return errors.New("stripePriceId is required")
	}
	if interval != "" && !interval.Valid() {
		return errors.New("invalid billingInterval")
	}
	if currency != "" && !currency.Valid() {
		return errors.New("invalid currency")
	}
	if quotaID == uuid.Nil {
		return errors.New("quotaId is required")
	}
	return nil
}
