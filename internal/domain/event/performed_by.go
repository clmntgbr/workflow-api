package event

import "github.com/google/uuid"

type PerformedBy struct {
	PerformedByUserID *string `json:"performedByUserId,omitempty"`
}

func OptionalUserIDString(id uuid.UUID) *string {
	if id == uuid.Nil {
		return nil
	}
	s := id.String()
	return &s
}
