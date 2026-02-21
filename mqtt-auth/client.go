package mqttAuth

import (
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const UsersTopic = "users"

type Client struct {
	mqtt.Client
	userID string
}

func NewClient(options *ClientOptions) *Client {
	client := mqtt.NewClient(options.ClientOptions)

	return &Client{Client: client, userID: options.userID}
}

func (c *Client) userTopic(topic string) string {
	return fmt.Sprintf("%s/%s/%s", UsersTopic, c.userID, topic)
}

func (c *Client) Connect() error {
	const op = "mqtt-auth.client.Connect"

	if token := c.Client.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("%s: failed to connect: %w", op, token.Error())
	}

	return nil
}

func (c *Client) Publish(
	topic string,
	qos byte,
	retained bool,
	payload any,
) error {
	const op = "mqtt-auth.client.Publish"

	userTopic := c.userTopic(topic)

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
	topic string,
	qos byte,
	callback mqtt.MessageHandler,
) error {
	const op = "mqtt-auth.client.Subscribe"

	userTopic := c.userTopic(topic)

	if token := c.Client.Subscribe(userTopic, qos, callback); token.Wait() && token.Error() != nil {
		return fmt.Errorf("%s: failed to subscribe on topic %q: %w", op, userTopic, token.Error())
	}

	return nil
}

func (c *Client) SubscribeMultiple(filters map[string]byte, callback mqtt.MessageHandler) error {
	const op = "mqtt-auth.client.SubscribeMultiple"

	userFilters := make(map[string]byte, len(filters))
	for topic, filter := range filters {
		userFilters[c.userTopic(topic)] = filter
	}

	if token := c.Client.SubscribeMultiple(
		userFilters,
		callback,
	); token.Wait() &&
		token.Error() != nil {
		return fmt.Errorf("%s: failed to subscribe: %w", op, token.Error())
	}

	return nil
}

func (c *Client) Unsubscribe(topics ...string) error {
	const op = "mqtt-auth.client.Unsubscribe"

	userTopics := make([]string, len(topics))
	for i, topic := range topics {
		userTopics[i] = c.userTopic(topic)
	}

	if token := c.Client.Unsubscribe(userTopics...); token.Wait() && token.Error() != nil {
		return fmt.Errorf("%s: failed to unsubscribe: %w", op, token.Error())
	}

	return nil
}

func (c *Client) AddRoute(topic string, callback mqtt.MessageHandler) {
	c.Client.AddRoute(c.userTopic(topic), callback)
}
