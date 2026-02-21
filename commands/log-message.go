package commands

import (
	"time"
)

const (
	DoneCode = iota
	CommandErrorCode
	InternalErrorCode
)

type LogMessageData struct {
	Command     string    `json:"command"`
	ReceivedAt  time.Time `json:"receivedAt"`
	CompletedAt time.Time `json:"completedAt"`
	Code        int       `json:"code"`
	Message     string    `json:"message,omitempty"`
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

func (m *LogMessage) Done() *LogMessage {
	m.Data.Code = DoneCode
	return m
}

func (m *LogMessage) CommandFailed(err *CommandError) *LogMessage {
	m.Data.Code = CommandErrorCode
	m.Data.Message = err.Error()
	return m
}

func (m *LogMessage) Internal() *LogMessage {
	m.Data.Code = InternalErrorCode
	return m
}
