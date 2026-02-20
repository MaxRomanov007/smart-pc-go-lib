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
	CompletedAt time.Time `json:"completeAt"`
	Code        int       `json:"code"`
	Message     string    `json:"message,omitempty"`
}

type LogMessage struct {
	Type string         `json:"type"`
	Data LogMessageData `json:"data"`
}

func newResponse(
	command, message, messageType string,
	code int,
	receivedAt, completedAt time.Time,
) *LogMessage {
	return &LogMessage{
		Type: messageType,
		Data: LogMessageData{
			Command:     command,
			ReceivedAt:  receivedAt,
			CompletedAt: completedAt,
			Message:     message,
			Code:        code,
		},
	}
}

func Done(command, messageType string, receivedAt, completedAt time.Time) *LogMessage {
	return newResponse(command, "", messageType, DoneCode, receivedAt, completedAt)
}

func CommandFailed(
	command, messageType string,
	err *CommandError,
	receivedAt, completedAt time.Time,
) *LogMessage {
	return newResponse(command, err.Error(), messageType, CommandErrorCode, receivedAt, completedAt)
}

func Internal(command, messageType string, receivedAt, completedAt time.Time) *LogMessage {
	return newResponse(
		command,
		"internal error",
		messageType,
		InternalErrorCode,
		receivedAt,
		completedAt,
	)
}
