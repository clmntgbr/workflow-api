package project

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	domainproject "go-api/internal/domain/project"
)

type ProjectMemberRemovedHandler struct{}

func NewProjectMemberRemovedHandler() *ProjectMemberRemovedHandler {
	return &ProjectMemberRemovedHandler{}
}

func (h *ProjectMemberRemovedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainproject.ProjectMemberRemoved
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	log.Printf(
		"event handled %s eventId=%s projectId=%s userId=%s",
		domainproject.EventTypeProjectMemberRemoved,
		evt.ID,
		evt.ProjectID,
		evt.UserID,
	)
	return nil
}
