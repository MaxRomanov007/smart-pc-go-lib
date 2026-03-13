package mqttAuth

import (
	"context"
	"fmt"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/packets"
	"github.com/eclipse/paho.golang/paho"
)

type Connection struct {
	*autopaho.ConnectionManager
	topicFactory *TopicFactory
	clientConfig *ClientConfig
}

func NewConnection(ctx context.Context, cfg *ClientConfig) (*Connection, error) {
	const op = "mqtt-auth.connection.NewConnection"

	connectionManager, err := newConnectionManager(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create connection manager: %w", op, err)
	}

	return &Connection{
		ConnectionManager: connectionManager,
		topicFactory:      cfg.topicFactory,
		clientConfig:      cfg,
	}, nil
}

func newConnectionManager(
	ctx context.Context,
	cfg *ClientConfig,
) (*autopaho.ConnectionManager, error) {
	const op = "mqtt-auth.connection.newConnectionManager"

	connection, err := autopaho.NewConnection(ctx, cfg.ClientConfig)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create connection: %w", op, err)
	}
	if err = connection.AwaitConnection(ctx); err != nil {
		return nil, fmt.Errorf("%s: failed to await connection: %w", op, err)
	}

	return connection, nil
}

func (c *Connection) Subscribe(ctx context.Context, s *paho.Subscribe) (*paho.Suback, error) {
	const op = "mqtt-auth.connection.Subscribe"

	for i := 0; i < len(s.Subscriptions); i++ {
		s.Subscriptions[i].Topic = c.topicFactory.UserTopic(s.Subscriptions[i].Topic)
	}

	ack, err := c.ConnectionManager.Subscribe(ctx, s)
	if err != nil {
		if ack != nil && ack.Reasons != nil && len(ack.Reasons) > 0 {
			return nil, fmt.Errorf(
				"%s: failed to subscribe (reason %d): %w",
				op,
				ack.Reasons[0],
				err,
			)
		}
		return ack, fmt.Errorf("%s: failed to subscribe: %w", op, err)
	}

	return ack, nil
}

func (c *Connection) Unsubscribe(ctx context.Context, u *paho.Unsubscribe) (*paho.Unsuback, error) {
	const op = "mqtt-auth.connection.Subscribe"

	for i := 0; i < len(u.Topics); i++ {
		u.Topics[i] = c.topicFactory.UserTopic(u.Topics[i])
	}

	ack, err := c.ConnectionManager.Unsubscribe(ctx, u)
	if err != nil {
		return ack, fmt.Errorf("%s: failed to unsubscribe: %w", op, err)
	}

	return ack, nil
}

func (c *Connection) Publish(ctx context.Context, p *paho.Publish) (*paho.PublishResponse, error) {
	const op = "mqtt-auth.connection.Subscribe"

	p.Topic = c.topicFactory.UserTopic(p.Topic)

	ack, err := c.ConnectionManager.Publish(ctx, p)
	if err == nil {
		return ack, nil
	}

	if ack == nil || ack.ReasonCode != packets.PubackNotAuthorized {
		return ack, fmt.Errorf("%s: failed to publish: %w", op, err)
	}

	if err := c.Renew(ctx); err != nil {
		return ack, fmt.Errorf("%s: failed to renew connection: %w", op, err)
	}

	ack, err = c.ConnectionManager.Publish(ctx, p)
	if err != nil {
		return ack, fmt.Errorf("%s: failed to publish after renew: %w", op, err)
	}

	return ack, nil
}

func (c *Connection) PublishViaQueue(ctx context.Context, p *autopaho.QueuePublish) error {
	const op = "mqtt-auth.connection.Subscribe"

	p.Topic = c.topicFactory.UserTopic(p.Topic)

	err := c.ConnectionManager.PublishViaQueue(ctx, p)
	if err != nil {
		return fmt.Errorf("%s: failed to publish: %w", op, err)
	}

	return nil
}

func (c *Connection) Renew(ctx context.Context) error {
	const op = "mqtt-auth.connection.Renew"

	if err := c.ConnectionManager.Disconnect(ctx); err != nil {
		return fmt.Errorf("%s: failed to disconnect: %w", op, err)
	}

	connectionManager, err := newConnectionManager(ctx, c.clientConfig)
	if err != nil {
		return fmt.Errorf("%s: failed to create connection manager: %w", op, err)
	}

	c.ConnectionManager = connectionManager
	return nil
}
