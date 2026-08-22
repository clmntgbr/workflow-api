package project

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	domainproject "go-api/internal/domain/project"
)

type ProjectMemberAddedHandler struct{}

func NewProjectMemberAddedHandler() *ProjectMemberAddedHandler {
	return &ProjectMemberAddedHandler{}
}

func (h *ProjectMemberAddedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainproject.ProjectMemberAdded
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	log.Printf(
		"event handled %s eventId=%s projectId=%s userId=%s",
		domainproject.EventTypeProjectMemberAdded,
		evt.ID,
		evt.ProjectID,
		evt.UserID,
	)
	return nil
}
