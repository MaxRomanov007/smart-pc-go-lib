package commands

import (
	"time"
)

const (
	StatusOK            = "ok"
	StatusCommandError  = "command-error"
	StatusInternalError = "internal-error"
)

type LogMessageData struct {
	Command     string    `json:"command"`
	ReceivedAt  time.Time `json:"receivedAt"`
	CompletedAt time.Time `json:"completedAt"`
	Status      string    `json:"status"`
	Error       string    `json:"error,omitempty"`
}

type LogMessage struct {
	Type string         `json:"type"`
	Data LogMessageData `json:"data"`
}

func NewLogMessage(
	command, messageType string,
	receivedAt, completedAt time.Time,
) *LogMessage {
	return &LogMessage{
		Type: messageType,
		Data: LogMessageData{
			Command:     command,
			ReceivedAt:  receivedAt,
			CompletedAt: completedAt,
		},
	}
}

func (m *LogMessage) OK() *LogMessage {
	m.Data.Status = StatusOK
	return m
}

func (m *LogMessage) CommandFailed(err *CommandError) *LogMessage {
	m.Data.Status = StatusCommandError
	m.Data.Error = err.Error()
	return m
}

func (m *LogMessage) Internal() *LogMessage {
	m.Data.Status = StatusInternalError
	return m
}
