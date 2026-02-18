package commands

import (
	"errors"
	"fmt"
	"time"

	"github.com/MaxRomanov007/smart-pc-go-lib/authorization"
)

type StartOptions struct {
	Auth              *authorization.Auth
	Topic             string
	MessageType       string
	LogTopic          string
	LogMessageType    string
	ReconnectDelay    time.Duration
	ReconnectAttempts int
}

func (o *StartOptions) check() error {
	errs := make([]error, 0, 3)

	if o.Auth == nil {
		errs = append(errs, errors.New("authorization required"))
	}
	if o.Topic == "" {
		errs = append(errs, errors.New("topic required"))
	}
	if o.MessageType == "" {
		errs = append(errs, errors.New("message type required"))
	}
	if o.LogTopic == "" {
		errs = append(errs, errors.New("log topic type required"))
	}
	if o.LogMessageType == "" {
		errs = append(errs, errors.New("log message type required"))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (o *StartOptions) UserTopic(userID string) string {
	return fmt.Sprintf("users/%s/%s", userID, o.Topic)
}
