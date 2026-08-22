package subscription

type Status string

const (
	StatusActive    Status = "active"
	StatusInactive  Status = "inactive"
	StatusCancelled Status = "cancelled"
	StatusPending   Status = "pending"
	StatusPastDue   Status = "past_due"
)

func (s Status) Valid() bool {
	switch s {
	case StatusActive, StatusInactive, StatusCancelled, StatusPending, StatusPastDue:
		return true
	default:
		return false
	}
}
