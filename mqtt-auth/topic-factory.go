package mqttAuth

import "fmt"

const UsersTopic = "users"

type TopicFactory struct {
	userID string
}

func NewTopicFactory(userID string) *TopicFactory {
	return &TopicFactory{userID: userID}
}

func (f *TopicFactory) UserTopic(topic string) string {
	return fmt.Sprintf("%s/%s/%s", UsersTopic, f.userID, topic)
}
