package mqttAuth

import "github.com/eclipse/paho.golang/paho"

type Router struct {
	*paho.StandardRouter
	topicFactory *TopicFactory
}

func NewRouter(topicFactory *TopicFactory) *Router {
	router := paho.NewStandardRouter()

	return &Router{StandardRouter: router, topicFactory: topicFactory}
}

func (r *Router) RegisterHandler(topic string, h paho.MessageHandler) {
	userTopic := r.topicFactory.UserTopic(topic)
	r.StandardRouter.RegisterHandler(userTopic, h)
}

func (r *Router) UnregisterHandler(topic string) {
	userTopic := r.topicFactory.UserTopic(topic)
	r.StandardRouter.UnregisterHandler(userTopic)
}
