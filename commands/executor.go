package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/MaxRomanov007/smart-pc-go-lib/authorization"
	"github.com/MaxRomanov007/smart-pc-go-lib/domain/models/message"
	"github.com/MaxRomanov007/smart-pc-go-lib/logger/sl"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type CommandFunc func(context.Context, *message.Message) error

type Executor struct {
	commands       map[string]CommandFunc
	defaultCommand CommandFunc
	client         mqtt.Client
	commandTopic   string
}

func NewExecutor() *Executor {
	return &Executor{
		commands:       make(map[string]CommandFunc),
		defaultCommand: nil,
	}
}

func (e *Executor) Set(name string, command CommandFunc) {
	e.commands[name] = command
}

func (e *Executor) SetDefault(command CommandFunc) {
	e.defaultCommand = command
}

func (e *Executor) Connect(ctx context.Context, opts *ConnectOptions) error {
	const op = "commands.executor.Start"

	if err := opts.check(); err != nil {
		return fmt.Errorf("%s: options validate failed: %w", op, err)
	}

	token, err := opts.Auth.Token(ctx)
	if err != nil {
		return fmt.Errorf("%s: failed to get token: %w", op, err)
	}
	userinfo, err := opts.Auth.FetchUserInfo(ctx)
	if err != nil {
		return fmt.Errorf("%s: fetch user info failed: %w", op, err)
	}
	e.commandTopic = opts.UserTopic(userinfo.Sub)

	opts.MQTTOptions.SetUsername(userinfo.Sub)
	opts.MQTTOptions.SetPassword(token)
	opts.MQTTOptions.SetCleanSession(false)
	opts.MQTTOptions.SetReconnectingHandler(reconnectHandler(ctx, opts.Log, opts.Auth))

	client := mqtt.NewClient(opts.MQTTOptions)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("%s: failed to connect: %w", op, token.Error())
	}
	e.client = client

	if token := e.client.Subscribe(
		e.commandTopic,
		1,
		e.messageHandler(
			ctx,
			opts.Log,
			opts.MessageType,
			opts.UserLogTopic(userinfo.Sub),
			opts.LogMessageType,
		),
	); token.Wait() &&
		token.Error() != nil {
		return fmt.Errorf("%s: failed to subscribe: %w", op, token.Error())
	}

	return nil
}

func (e *Executor) Disconnect() error {
	const op = "commands.executor.Disconnect"

	if token := e.client.Unsubscribe(e.commandTopic); token.Wait() && token.Error() != nil {
		return fmt.Errorf(
			"%s: failed to unsubscribe from topic %q: %w",
			op,
			e.commandTopic,
			token.Error(),
		)
	}

	e.client.Disconnect(250)

	return nil
}

func reconnectHandler(
	ctx context.Context,
	log *slog.Logger,
	auth *authorization.Auth,
) mqtt.ReconnectHandler {
	return func(_ mqtt.Client, opts *mqtt.ClientOptions) {
		const op = "commands.executor.reconnect"

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

func (e *Executor) messageHandler(
	ctx context.Context,
	log *slog.Logger,
	messageType, logTopic, logMessageType string,
) mqtt.MessageHandler {
	return func(client mqtt.Client, mqttMessage mqtt.Message) {
		const op = "commands.executor.messageHandler"

		log := log.With(sl.Op(op), sl.MsgId(mqttMessage))
		log.Debug("received message")

		receivedAt := time.Now()

		msg := new(message.Message)
		if err := json.Unmarshal(mqttMessage.Payload(), msg); err != nil {
			log.Error("failed to unmarshal payload", sl.Err(err))
			return
		}

		if msg.Type != messageType {
			log.Debug("invalid message type, skipping")
			return
		}

		log.Info(
			"received command",
			slog.String("command", msg.Data.Command),
		)

		handler := e.getCommand(msg.Data.Command)
		if handler == nil {
			log.Warn("handler not found, skipping")
			return
		}

		err := handler(ctx, msg)

		completedAt := time.Now()

		if err != nil {
			if commandErr, ok := errors.AsType[*CommandError](err); ok {
				log.Info("command error", sl.Err(commandErr))

				if err := sendLog(
					client,
					logTopic,
					CommandFailed(
						msg.Data.Command,
						logMessageType,
						commandErr,
						receivedAt,
						completedAt,
					),
				); err != nil {
					log.Warn("failed to send command error log", sl.Err(err))
				}
				return
			}

			log.Error("failed to handle message", sl.Err(err))
			if err := sendLog(
				client,
				logTopic,
				Internal(
					msg.Data.Command,
					logMessageType,
					receivedAt,
					completedAt,
				),
			); err != nil {
				log.Warn("failed to send internal error log", sl.Err(err))
			}
			return
		}

		if err := sendLog(
			client,
			logTopic,
			Done(msg.Data.Command, logMessageType, receivedAt, completedAt),
		); err != nil {
			log.Warn("failed to send done log", sl.Err(err))
			return
		}
	}
}

func (e *Executor) getCommand(key string) CommandFunc {
	if command, ok := e.commands[key]; ok {
		return command
	}

	return e.defaultCommand
}

func sendLog(client mqtt.Client, topic string, resp *LogMessage) error {
	const op = "commands.response.Send"

	data, err := json.Marshal(*resp)
	if err != nil {
		return fmt.Errorf("%s: failed to marshal json: %w", op, err)
	}

	if token := client.Publish(topic, 0, false, data); token.Wait() && token.Error() != nil {
		return fmt.Errorf("%s: failed to send log: %w", op, token.Error())
	}

	return nil
}
