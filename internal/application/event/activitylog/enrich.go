package activitylog

import (
	"context"
	"encoding/json"
	"strings"

	domainactivitylog "go-api/internal/domain/activitylog"

	"github.com/google/uuid"
)

func (h *RecordHandler) enrichHints(ctx context.Context, entry *domainactivitylog.Entry, hints *messageHints) error {
	if entry.WorkflowID != nil {
		if hints.WorkflowName == "" {
			workflow, err := h.workflowRead.FindByID(ctx, *entry.WorkflowID)
			if err != nil {
				return err
			}
			if workflow != nil {
				hints.WorkflowName = workflow.Name
				if hints.WorkflowStatus == "" {
					hints.WorkflowStatus = string(workflow.Status)
				}
			}
		}
	}

	if entry.StepID != nil && hints.StepName == "" {
		if err := h.fillStepHints(ctx, *entry.StepID, hints); err != nil {
			return err
		}
	}

	if hints.SourceStepID != uuid.Nil && hints.SourceStepName == "" {
		if err := h.fillStepName(ctx, hints.SourceStepID, &hints.SourceStepName); err != nil {
			return err
		}
	}
	if hints.TargetStepID != uuid.Nil && hints.TargetStepName == "" {
		if err := h.fillStepName(ctx, hints.TargetStepID, &hints.TargetStepName); err != nil {
			return err
		}
	}

	if err := h.enrichActorName(ctx, entry, hints); err != nil {
		return err
	}

	return nil
}

func (h *RecordHandler) enrichActorName(ctx context.Context, entry *domainactivitylog.Entry, hints *messageHints) error {
	if entry.ActorUserID == nil {
		return nil
	}
	user, err := h.userRead.FindByID(ctx, *entry.ActorUserID)
	if err != nil {
		return err
	}
	if user == nil {
		return nil
	}
	hints.ActorUserName = userDisplayName(user.FirstName, user.LastName, user.Email)
	return nil
}

func userDisplayName(firstName, lastName, email string) string {
	name := strings.TrimSpace(strings.TrimSpace(firstName) + " " + strings.TrimSpace(lastName))
	if name != "" {
		return name
	}
	return strings.TrimSpace(email)
}

func applyPerformedByFromPayload(entry *domainactivitylog.Entry, payload []byte) {
	if entry.ActorUserID != nil {
		return
	}
	var meta struct {
		PerformedByUserID *string `json:"performedByUserId"`
	}
	if err := json.Unmarshal(payload, &meta); err != nil {
		return
	}
	if meta.PerformedByUserID == nil || *meta.PerformedByUserID == "" {
		return
	}
	userID, err := uuid.Parse(*meta.PerformedByUserID)
	if err != nil {
		return
	}
	entry.ActorType = domainactivitylog.ActorUser
	entry.ActorUserID = &userID
}

func (h *RecordHandler) fillStepHints(ctx context.Context, stepID uuid.UUID, hints *messageHints) error {
	step, err := h.stepRead.FindByID(ctx, stepID)
	if err != nil {
		return err
	}
	if step == nil {
		return nil
	}
	hints.StepName = step.Name
	if hints.Method == "" {
		hints.Method = step.Method
	}
	if hints.URL == "" {
		hints.URL = step.URL
	}
	if hints.WorkflowName == "" {
		workflow, err := h.workflowRead.FindByID(ctx, step.WorkflowID)
		if err != nil {
			return err
		}
		if workflow != nil {
			hints.WorkflowName = workflow.Name
		}
	}
	return nil
}

func (h *RecordHandler) fillStepName(ctx context.Context, stepID uuid.UUID, name *string) error {
	step, err := h.stepRead.FindByID(ctx, stepID)
	if err != nil {
		return err
	}
	if step != nil {
		*name = step.Name
	}
	return nil
}
