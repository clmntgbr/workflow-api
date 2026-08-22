package project

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	domainproject "go-api/internal/domain/project"
)

type ProjectUpdatedHandler struct{}

func NewProjectUpdatedHandler() *ProjectUpdatedHandler {
	return &ProjectUpdatedHandler{}
}

func (h *ProjectUpdatedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainproject.ProjectUpdated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	log.Printf(
		"event handled %s eventId=%s projectId=%s name=%s",
		domainproject.EventTypeProjectUpdated,
		evt.ID,
		evt.ProjectID,
		evt.Name,
	)
	return nil
}
