package workflowrun

import "fmt"

type TriggeredBy string

const (
	TriggeredByUser     TriggeredBy = "user"
	TriggeredBySchedule TriggeredBy = "schedule"
	TriggeredByWebhook  TriggeredBy = "webhook"
	TriggeredByAPI      TriggeredBy = "api"
	TriggeredByCLI      TriggeredBy = "cli"
)

func (t TriggeredBy) Valid() bool {
	switch t {
	case TriggeredByUser, TriggeredBySchedule, TriggeredByWebhook, TriggeredByAPI, TriggeredByCLI:
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
