package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/MaxRomanov007/smart-pc-go-lib/domain/models/message"
	"github.com/MaxRomanov007/smart-pc-go-lib/logger/sl"
	mqttAuth "github.com/MaxRomanov007/smart-pc-go-lib/mqtt-auth"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type CommandFunc func(context.Context, *message.Message) error

type Executor struct {
	commands       map[string]CommandFunc
	defaultCommand CommandFunc
	client         *mqttAuth.Client
	commandTopic   string
}

func NewExecutor(client *mqttAuth.Client) *Executor {
	return &Executor{
		commands:       make(map[string]CommandFunc),
		defaultCommand: nil,
		client:         client,
	}
}

func (e *Executor) Set(name string, command CommandFunc) {
	e.commands[name] = command
}

func (e *Executor) SetDefault(command CommandFunc) {
	e.defaultCommand = command
}

func (e *Executor) StartListen(ctx context.Context, opts *StartListenOptions) error {
	const op = "commands.executor.StartListen"

	if err := opts.check(); err != nil {
		return fmt.Errorf("%s: options validate failed: %w", op, err)
	}

	e.commandTopic = opts.CommandTopic

	if err := e.client.Subscribe(
		e.commandTopic,
		1,
		e.messageHandler(
			ctx,
			opts.Log,
			opts.CommandMessageType,
			opts.LogTopic,
			opts.LogMessageType,
		),
	); err != nil {
		return fmt.Errorf("%s: failed to subscribe: %w", op, err)
	}

	return nil
}

func (e *Executor) StopListen() error {
	const op = "commands.executor.StopListen"

	if err := e.client.Unsubscribe(e.commandTopic); err != nil {
		return fmt.Errorf(
			"%s: failed to unsubscribe from topic %q: %w",
			op,
			e.commandTopic,
			err,
		)
	}

	return nil
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
		logMessage := NewLogMessage(msg.Data.Command, logMessageType, receivedAt, completedAt)

		if err == nil {
			if err := e.sendLog(logTopic, logMessage.Done()); err != nil {
				log.Warn("failed to send done log", sl.Err(err))
				return
			}
			return
		}

		if commandErr, ok := errors.AsType[*CommandError](err); ok {
			log.Info("command error", sl.Err(commandErr))

			if err := e.sendLog(logTopic, logMessage.CommandFailed(commandErr)); err != nil {
				log.Warn("failed to send command error log", sl.Err(err))
			}
			return
		}

		log.Error("failed to handle message", sl.Err(err))
		if err := e.sendLog(logTopic, logMessage.Internal()); err != nil {
			log.Warn("failed to send internal error log", sl.Err(err))
		}
	}
}

func (e *Executor) getCommand(key string) CommandFunc {
	if command, ok := e.commands[key]; ok {
		return command
	}

	return e.defaultCommand
}

func (e *Executor) sendLog(topic string, resp *LogMessage) error {
	const op = "commands.response.sendLog"

	data, err := json.Marshal(*resp)
	if err != nil {
		return fmt.Errorf("%s: failed to marshal json: %w", op, err)
	}

	if err := e.client.Publish(topic, 0, false, data); err != nil {
		return fmt.Errorf("%s: failed to send log: %w", op, err)
	}

	return nil
}
