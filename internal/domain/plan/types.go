package plan

import "fmt"

const FreePlanSlug = "free"

type BillingInterval string

const (
	BillingIntervalMonth BillingInterval = "month"
	BillingIntervalYear  BillingInterval = "year"
)

func (i BillingInterval) Valid() bool {
	switch i {
	case BillingIntervalMonth, BillingIntervalYear:
		return true
	default:
		return false
	}
}

func ParseBillingInterval(value string) (BillingInterval, error) {
	if value == "" {
		return BillingIntervalMonth, nil
	}
	i := BillingInterval(value)
	if !i.Valid() {
		return "", fmt.Errorf("invalid billing interval %q", value)
	}
	return i, nil
}

type Currency string

const (
	CurrencyEUR Currency = "EUR"
	CurrencyUSD Currency = "USD"
)

func (c Currency) Valid() bool {
	switch c {
	case CurrencyEUR, CurrencyUSD:
		return true
	default:
		return false
	}
}

func ParseCurrency(value string) (Currency, error) {
	if value == "" {
		return CurrencyEUR, nil
	}
	c := Currency(value)
	if !c.Valid() {
		return "", fmt.Errorf("invalid currency %q", value)
	}
	return c, nil
}
