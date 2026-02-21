package mqttAuth

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/MaxRomanov007/smart-pc-go-lib/authorization"
	"github.com/MaxRomanov007/smart-pc-go-lib/logger/sl"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type ClientOptions struct {
	*mqtt.ClientOptions
	topicFactory *TopicFactory
}

func NewClientOptions(
	ctx context.Context,
	log *slog.Logger,
	auth *authorization.Auth,
) (*ClientOptions, error) {
	const op = "mqtt-auth.client-options.NewClientOptions"

	options := mqtt.NewClientOptions()

	token, err := auth.Token(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get token: %w", op, err)
	}
	userinfo, err := auth.FetchUserInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: fetch user info failed: %w", op, err)
	}

	options.SetUsername(userinfo.Sub)
	options.SetPassword(token)
	options.SetCleanSession(false)
	options.SetReconnectingHandler(reconnectHandler(ctx, log, auth))

	return &ClientOptions{ClientOptions: options, topicFactory: NewTopicFactory(userinfo.Sub)}, nil
}

func reconnectHandler(
	ctx context.Context,
	log *slog.Logger,
	auth *authorization.Auth,
) mqtt.ReconnectHandler {
	return func(_ mqtt.Client, opts *mqtt.ClientOptions) {
		const op = "commands.executor.reconnectHandler"

		log := log.With(sl.Op(op))

		log.Debug("reconnecting")

		token, err := auth.Token(ctx)
		if err != nil {
			log.Warn("failed to fetch token", sl.Err(err))
			return
		}

		opts.SetPassword(token)
	}
}

func (o *ClientOptions) SetWill(
	topic string,
	payload string,
	qos byte,
	retained bool,
) *mqtt.ClientOptions {
	userTopic := o.topicFactory.UserTopic(topic)
	return o.ClientOptions.SetWill(userTopic, payload, qos, retained)
}
