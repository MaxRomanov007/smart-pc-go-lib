package models

import "github.com/google/uuid"

type Pc struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CanPowerOn  bool      `json:"canPowerOn"`

	Commands []Command `json:"commands,omitempty"`
}
