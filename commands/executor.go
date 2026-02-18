package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/MaxRomanov007/smart-pc-go-lib/authorization"
	"github.com/MaxRomanov007/smart-pc-go-lib/domain/models/message"
	"github.com/MaxRomanov007/smart-pc-go-lib/logger/sl"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gorilla/websocket"
)

const connectionCheckTimeout = time.Second

type CommandFunc func(context.Context, *message.Message) error

type Executor struct {
	commands       map[string]CommandFunc
	defaultCommand CommandFunc
	clientOptions  *mqtt.ClientOptions
}

func NewExecutor(opts *mqtt.ClientOptions) (*Executor, error) {
	return &Executor{
		commands:       make(map[string]CommandFunc),
		defaultCommand: nil,
		clientOptions:  opts,
	}, nil
}

func (e *Executor) Set(name string, command CommandFunc) {
	e.commands[name] = command
}

func (e *Executor) SetDefault(command CommandFunc) {
	e.defaultCommand = command
}

func (e *Executor) Start(ctx context.Context, l *slog.Logger, opts *StartOptions) error {
	const op = "commands.executor.Start"

	log := l.With(sl.Op(op))

	if err := opts.check(); err != nil {
		return fmt.Errorf("%s: options validate failed: %w", op, err)
	}

	failingsCount := 0
	for failingsCount <= opts.ReconnectAttempts {
		token, authErr := opts.Auth.Token(ctx)
		if authErr == nil {
			log.Debug("listening channel")

			userinfo, err := opts.Auth.FetchUserInfo(ctx)
			if err != nil {
				return fmt.Errorf("%s: failed to fetch user info: %w", op, err)
			}

			e.clientOptions.SetUsername(userinfo.Sub)
			e.clientOptions.SetPassword(token)
			e.clientOptions.SetReconnectingHandler(
				func(client mqtt.Client, options *mqtt.ClientOptions) {
				},
			)
			err = e.listen(
				ctx,
				mqtt.NewClient(e.clientOptions),
				log,
				opts.UserTopic(userinfo.Sub),
				opts.MessageType,
				opts.LogTopic,
				opts.LogMessageType,
			)
			if err == nil {
				failingsCount = 0
				continue
			}

			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return fmt.Errorf("%s: context canceled: %w", op, err)
			}

			log.Warn("failed to listen channel", sl.Err(err))
		}
		if authErr != nil {
			log.Warn("failed to get token", sl.Err(authErr))
		}

		failingsCount++

		if failingsCount <= opts.ReconnectAttempts {
			log.Warn("reconnecting", slog.Int("attempt", failingsCount))
			time.Sleep(opts.ReconnectDelay)
		}
	}

	return nil
}

func isConnectionClosedError(err error) bool {
	return errors.Is(err, websocket.ErrCloseSent) ||
		websocket.IsCloseError(
			err,
			websocket.CloseAbnormalClosure,
			websocket.CloseNoStatusReceived,
		) ||
		strings.Contains(err.Error(), "closed network connection")
}

func isSyntaxError(err error) bool {
	var jsonErr *json.SyntaxError
	return errors.As(err, &jsonErr)
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

func (e *Executor) listen(
	ctx context.Context,
	client mqtt.Client,
	l *slog.Logger,
	topic, messageType, logTopic, logMessageType string,
) error {
	const op = "commands.executor.listen"

	log := l.With(sl.Op(op))

	client.IsConnected()

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("%s: failed to connect: %w", op, token.Error())
	}
	defer client.Disconnect(250)

	if token := client.Subscribe(
		topic,
		1,
		e.messageHandler(ctx, log, messageType, logTopic, logMessageType),
	); token.Wait() &&
		token.Error() != nil {
		return fmt.Errorf("%s: failed to subscribe: %w", op, token.Error())
	}
	defer client.Unsubscribe(topic)

	ticker := time.NewTicker(connectionCheckTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !client.IsConnected() {
				return nil
			}
		case <-ctx.Done():
			return fmt.Errorf("%s: context canceled: %w", op, ctx.Err())
		}
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

		handler := e.getHandler(msg.Data.Command)
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

func (e *Executor) getHandler(key string) CommandFunc {
	if handler, ok := e.commands[key]; ok {
		return handler
	}

	return e.defaultCommand
}

func sendLog(client mqtt.Client, topic string, resp *Response) error {
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
