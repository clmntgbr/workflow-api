package realtime

import (
	"context"
	"fmt"
	"log"

	"go-api/internal/application/messaging"
	"go-api/internal/domain/port"
	domainproject "go-api/internal/domain/project"

	"github.com/google/uuid"
)

// Publisher pushes domain events to the realtime transport. Event handlers own
// the decoding of their own payloads and delegate the delivery here, so the
// retry classification and logging stay identical across every entity.
type Publisher struct {
	transport   port.RealtimePublisher
	projectRepo domainproject.ProjectReadRepository
}

// NewPublisher builds a publisher. projectRepo may be nil for handlers that
// only ever publish to a single user.
func NewPublisher(
	transport port.RealtimePublisher,
	projectRepo domainproject.ProjectReadRepository,
) *Publisher {
	return &Publisher{transport: transport, projectRepo: projectRepo}
}

// ToProjectMembers resolves the project and delivers the event to every member.
//
// A malformed project id or a missing project is non-retryable: replaying the
// message cannot fix either. A repository failure is retryable.
func (p *Publisher) ToProjectMembers(
	ctx context.Context,
	entity string,
	action string,
	projectIDRaw string,
	payload any,
) error {
	projectID, err := uuid.Parse(projectIDRaw)
	if err != nil {
		return messaging.NonRetryable(err)
	}

	project, err := p.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return messaging.Retryable(err)
	}
	if project == nil {
		return messaging.NonRetryable(fmt.Errorf("project not found for %s realtime publish", entity))
	}

	return p.ToMembers(ctx, entity, action, projectIDRaw, project.MemberIDs, payload)
}

// ToMembers delivers the event to an already-resolved set of members. Use it
// when the event itself carries the recipient list.
func (p *Publisher) ToMembers(
	ctx context.Context,
	entity string,
	action string,
	projectIDRaw string,
	memberIDs []uuid.UUID,
	payload any,
) error {
	eventType := EventType(entity, action)
	for _, memberID := range memberIDs {
		if err := p.transport.PublishToUser(ctx, memberID, eventType, payload); err != nil {
			log.Printf(
				"centrifugo publish failed type=%s projectId=%s userId=%s: %v",
				eventType, projectIDRaw, memberID, err,
			)
			return messaging.Retryable(err)
		}
		log.Printf(
			"centrifugo published type=%s projectId=%s userId=%s",
			eventType, projectIDRaw, memberID,
		)
	}
	return nil
}

// ToUsers delivers the event to each of the given users. Use it when the event
// itself carries the recipient list as raw ids.
func (p *Publisher) ToUsers(
	ctx context.Context,
	entity string,
	action string,
	userIDsRaw []string,
	payload any,
) error {
	for _, userIDRaw := range userIDsRaw {
		if err := p.ToUser(ctx, entity, action, userIDRaw, payload); err != nil {
			return err
		}
	}
	return nil
}

// ToUser delivers the event to a single user.
func (p *Publisher) ToUser(
	ctx context.Context,
	entity string,
	action string,
	userIDRaw string,
	payload any,
) error {
	userID, err := uuid.Parse(userIDRaw)
	if err != nil {
		return messaging.NonRetryable(err)
	}

	eventType := EventType(entity, action)
	if err := p.transport.PublishToUser(ctx, userID, eventType, payload); err != nil {
		log.Printf("centrifugo publish failed type=%s userId=%s: %v", eventType, userIDRaw, err)
		return messaging.Retryable(err)
	}
	log.Printf("centrifugo published type=%s userId=%s", eventType, userIDRaw)
	return nil
}
