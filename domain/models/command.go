package models

import "github.com/google/uuid"

type Command struct {
	ID          *uuid.UUID `json:"id,omitempty"`
	PcID        *uuid.UUID `json:"pcId,omitempty"`
	Name        string     `json:"name,omitempty"`
	Description string     `json:"description,omitempty"`

	Pc         *Pc                `json:"pc,omitempty"`
	Parameters []CommandParameter `json:"parameters,omitempty"`
	Logs       []PcLog            `json:"logs,omitempty"`
}
