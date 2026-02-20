package mqttAuth

import (
	"context"
	"fmt"

	"github.com/MaxRomanov007/smart-pc-go-lib/authorization"
	"github.com/MaxRomanov007/smart-pc-go-lib/domain/models/user"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const UsersTopic = "users"

type Client struct {
	mqtt.Client
	auth     *authorization.Auth
	userinfo *user.Info
}

func (c *Client) userTopic(ctx context.Context, topic string) (string, error) {
	const op = "mqtt-auth.client.userTopic"

	userinfo, err := c.getUserInfo(ctx)
	if err != nil {
		return "", fmt.Errorf("%s: failed to get userinfo: %w", op, err)
	}

	return fmt.Sprintf("%s/%s/%s", UsersTopic, userinfo.Sub, topic), nil
}

func (c *Client) getUserInfo(ctx context.Context) (*user.Info, error) {
	const op = "mqtt-auth.client.getUserInfo"

	if c.userinfo != nil {
		return c.userinfo, nil
	}

	userinfo, err := c.auth.FetchUserInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to fetch userinfo: %w", op, err)
	}

	c.userinfo = userinfo
	return userinfo, nil
}

func (c *Client) Connect() error {
	const op = "mqtt-auth.client.Connect"

	if token := c.Client.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("%s: failed to connect: %w", op, token.Error())
	}

	return nil
}

func (c *Client) Publish(
	ctx context.Context,
	topic string,
	qos byte,
	retained bool,
	payload any,
) error {
	const op = "mqtt-auth.client.Publish"

	userTopic, err := c.userTopic(ctx, topic)
	if err != nil {
		return fmt.Errorf("%s: failed to get user topic: %w", op, err)
	}

	if token := c.Client.Publish(
		userTopic,
		qos,
		retained,
		payload,
	); token.Wait() &&
		token.Error() != nil {
		return fmt.Errorf("%s: failed to publish %q: %w", op, userTopic, token.Error())
	}

	return nil
}

func (c *Client) Subscribe(
	ctx context.Context,
	topic string,
	qos byte,
	callback mqtt.MessageHandler,
) error {
	const op = "mqtt-auth.client.Subscribe"

	userTopic, err := c.userTopic(ctx, topic)
	if err != nil {
		return fmt.Errorf("%s: failed to get user topic: %w", op, err)
	}

	if token := c.Client.Subscribe(userTopic, qos, callback); token.Wait() && token.Error() != nil {
		return fmt.Errorf("%s: failed to subscribe on topic %q: %w", op, userTopic, token.Error())
	}

	return nil
}

//    SubscribeMultiple(filters map[string]byte, callback MessageHandler) Token
//    Unsubscribe(topics ...string) Token
//    AddRoute(topic string, callback MessageHandler)
