package project

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	domainproject "go-api/internal/domain/project"
)

type ProjectCreatedHandler struct{}

func NewProjectCreatedHandler() *ProjectCreatedHandler {
	return &ProjectCreatedHandler{}
}

func (h *ProjectCreatedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainproject.ProjectCreated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	log.Printf(
		"event handled %s eventId=%s projectId=%s name=%s",
		domainproject.EventTypeProjectCreated,
		evt.ID,
		evt.ProjectID,
		evt.Name,
	)
	return nil
}
