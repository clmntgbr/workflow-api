package workflowrun

import "fmt"

// TriggeredBy is how a workflow run was started.
type TriggeredBy string

const (
	TriggeredByUser     TriggeredBy = "user"
	TriggeredBySchedule TriggeredBy = "schedule"
	TriggeredByWebhook  TriggeredBy = "webhook"
	TriggeredByAPI      TriggeredBy = "api"
)

func (t TriggeredBy) Valid() bool {
	switch t {
	case TriggeredByUser, TriggeredBySchedule, TriggeredByWebhook, TriggeredByAPI:
		return true
	default:
		return false
	}
}

func ParseTriggeredBy(value string) (TriggeredBy, error) {
	t := TriggeredBy(value)
	if !t.Valid() {
		return "", fmt.Errorf("invalid workflow run triggeredBy %q", value)
	}
	return t, nil
}
