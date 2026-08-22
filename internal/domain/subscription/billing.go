package subscription

func MapBillingStatus(stripeStatus string) Status {
	switch stripeStatus {
	case "active", "trialing":
		return StatusActive
	case "past_due", "unpaid":
		return StatusPastDue
	case "canceled":
		return StatusCancelled
	case "incomplete", "incomplete_expired", "paused":
		return StatusInactive
	default:
		return StatusInactive
	}
}
