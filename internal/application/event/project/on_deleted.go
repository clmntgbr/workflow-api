package project

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	domainproject "go-api/internal/domain/project"
)

type ProjectDeletedHandler struct{}

func NewProjectDeletedHandler() *ProjectDeletedHandler {
	return &ProjectDeletedHandler{}
}

func (h *ProjectDeletedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainproject.ProjectDeleted
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	log.Printf(
		"event handled %s eventId=%s projectId=%s",
		domainproject.EventTypeProjectDeleted,
		evt.ID,
		evt.ProjectID,
	)
	return nil
}
