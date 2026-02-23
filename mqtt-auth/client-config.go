package mqttAuth

import (
	"context"
	"fmt"
	"net/url"

	"github.com/MaxRomanov007/smart-pc-go-lib/authorization"
	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
)

type ClientConfig struct {
	autopaho.ClientConfig
	topicFactory *TopicFactory
}

func NewClientConfig(
	ctx context.Context,
	auth *authorization.Auth,
) (*ClientConfig, error) {
	const op = "mqtt-auth.client-config.NewClientConfig"

	userinfo, err := auth.FetchUserInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to fetch user info: %w", op, err)
	}

	config := autopaho.ClientConfig{
		ConnectPacketBuilder: connectPacketBuilder(ctx, auth, userinfo.Sub),
	}

	return &ClientConfig{ClientConfig: config, topicFactory: NewTopicFactory(userinfo.Sub)}, nil
}

func NewClientConfigWithRouter(
	ctx context.Context,
	auth *authorization.Auth,
) (*ClientConfig, *Router, error) {
	const op = "mqtt-auth.client-config.NewClientConfigWithRouter"

	config, err := NewClientConfig(ctx, auth)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: failed to create client config: %w", op, err)
	}

	router := NewRouter(config.topicFactory)

	config.ClientConfig.ClientConfig.OnPublishReceived = []func(paho.PublishReceived) (bool, error){
		func(pr paho.PublishReceived) (bool, error) {
			router.Route(pr.Packet.Packet())
			return true, nil
		},
	}

	return config, router, nil
}

func connectPacketBuilder(
	ctx context.Context,
	auth *authorization.Auth,
	username string,
) func(*paho.Connect, *url.URL) (*paho.Connect, error) {
	return func(c *paho.Connect, u *url.URL) (*paho.Connect, error) {
		const op = "commands.client-config.connectPacketBuilder"

		token, err := auth.Token(ctx)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to fetch token: %w", op, err)
		}

		c.Username = username
		c.UsernameFlag = true
		c.Password = []byte(token)
		c.PasswordFlag = true

		return c, nil
	}
}

func (c *ClientConfig) SetWill(message *paho.WillMessage) {
	message.Topic = c.topicFactory.UserTopic(message.Topic)

	c.ClientConfig.WillMessage = message
}
