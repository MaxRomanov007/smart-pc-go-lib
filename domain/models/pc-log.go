package models

import (
	"time"

	"github.com/google/uuid"
)

type PcLog struct {
	ID          uuid.UUID  `json:"id"`
	PcID        *uuid.UUID `json:"pcId,omitempty"`
	CommandID   string     `json:"commandId"`
	ReceivedAt  time.Time  `json:"receivedAt"`
	CompletedAt time.Time  `json:"completedAt"`
	Status      string     `json:"status"`
	Error       *string    `json:"error,omitempty"`

	Pc      *Pc      `json:"pc,omitempty"`
	Command *Command `json:"command,omitempty"`
}
