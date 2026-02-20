package commands

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/MaxRomanov007/smart-pc-go-lib/authorization"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type ConnectOptions struct {
	Auth           *authorization.Auth
	Topic          string
	MessageType    string
	LogTopic       string
	LogMessageType string
	Log            *slog.Logger
	MQTTOptions    *mqtt.ClientOptions
}

func (o *ConnectOptions) check() error {
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
	if o.Log == nil {
		errs = append(errs, errors.New("log required"))
	}
	if o.MQTTOptions == nil {
		errs = append(errs, errors.New("mqtt option required"))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (o *ConnectOptions) UserTopic(userID string) string {
	return fmt.Sprintf("users/%s/%s", userID, o.Topic)
}

func (o *ConnectOptions) UserLogTopic(userID string) string {
	return fmt.Sprintf("users/%s/%s", userID, o.LogTopic)
}
