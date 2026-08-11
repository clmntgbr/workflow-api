package dto

import "encoding/json"

type ClerkEmailAddress struct {
	ID           string `json:"id"`
	EmailAddress string `json:"email_address"`
}

type ClerkUserCreated struct {
	ID             string              `json:"id" validate:"required"`
	FirstName      string              `json:"first_name" validate:"omitempty"`
	LastName       string              `json:"last_name" validate:"omitempty"`
	Banned         *bool               `json:"banned" validate:"required"`
	EmailAddresses []ClerkEmailAddress `json:"email_addresses"`
	Email          string              `json:"-"`
}

func (u *ClerkUserCreated) UnmarshalJSON(data []byte) error {
	type rawClerkUserCreated ClerkUserCreated
	var raw rawClerkUserCreated
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*u = ClerkUserCreated(raw)
	if len(u.EmailAddresses) > 0 {
		u.Email = u.EmailAddresses[0].EmailAddress
	}
	return nil
}

type ClerkUserUpdated struct {
	ID             string              `json:"id" validate:"required"`
	FirstName      string              `json:"first_name" validate:"omitempty"`
	LastName       string              `json:"last_name" validate:"omitempty"`
	Banned         *bool               `json:"banned" validate:"required"`
	EmailAddresses []ClerkEmailAddress `json:"email_addresses"`
	Email          string              `json:"-"`
}

func (u *ClerkUserUpdated) UnmarshalJSON(data []byte) error {
	type rawClerkUserUpdated ClerkUserUpdated
	var raw rawClerkUserUpdated
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*u = ClerkUserUpdated(raw)
	if len(u.EmailAddresses) > 0 {
		u.Email = u.EmailAddresses[0].EmailAddress
	}
	return nil
}

type ClerkUserDeleted struct {
	ID string `json:"id" validate:"required"`
}

type ClerkEvent struct {
	Type       string          `json:"type" validate:"required"`
	InstanceID string          `json:"instance_id" validate:"required"`
	Object     string          `json:"object" validate:"required"`
	Timestamp  int64           `json:"timestamp" validate:"required"`
	Data       json.RawMessage `json:"data"`
}
