package commands

import (
	"errors"
	"log/slog"

	commandMessage "github.com/MaxRomanov007/smart-pc-go-lib/domain/models/command-message"
)

type StartListenOptions struct {
	CommandTopic       string
	CommandMessageType string
	LogTopic           string
	LogTopicFunc       func(msg *commandMessage.Message) string
	LogMessageType     string
	Log                *slog.Logger
}

func (o *StartListenOptions) check() error {
	errs := make([]error, 0, 5)

	if o.CommandTopic == "" {
		errs = append(errs, errors.New("command topic required"))
	}
	if o.CommandMessageType == "" {
		errs = append(errs, errors.New("command message type required"))
	}
	if o.LogTopic == "" && o.LogTopicFunc == nil {
		errs = append(errs, errors.New("log topic or log topic func is required"))
	}
	if o.LogMessageType == "" {
		errs = append(errs, errors.New("log message type required"))
	}
	if o.Log == nil {
		errs = append(errs, errors.New("log required"))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
