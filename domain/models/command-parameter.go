package models

import "github.com/google/uuid"

type CommandParameter struct {
	ID          uuid.UUID `json:"id"`
	CommandID   uuid.UUID `json:"commandId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        int16     `json:"type"`

	Command *Command `json:"command,omitempty"`
}
