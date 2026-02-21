package commands

import (
	"errors"
	"log/slog"
)

type StartListenOptions struct {
	CommandTopic       string
	CommandMessageType string
	LogTopic           string
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
	if o.LogTopic == "" {
		errs = append(errs, errors.New("log topic type required"))
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
