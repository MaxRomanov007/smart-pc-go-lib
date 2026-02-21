package commands

import (
	"errors"
	"log/slog"

	mqttAuth "github.com/MaxRomanov007/smart-pc-go-lib/mqtt-auth"
)

type StartListenOptions struct {
	CommandTopic       string
	CommandMessageType string
	LogTopic           string
	LogMessageType     string
	Log                *slog.Logger
	Client             *mqttAuth.Client
}

func (o *StartListenOptions) check() error {
	errs := make([]error, 0, 6)

	if o.CommandTopic == "" {
		errs = append(errs, errors.New("topic required"))
	}
	if o.CommandMessageType == "" {
		errs = append(errs, errors.New("message type required"))
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
	if o.Client == nil {
		errs = append(errs, errors.New("mqtt client required"))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
