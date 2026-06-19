package pubsub

import (
	"context"

	"github.com/ThreeDotsLabs/watermill"
	wmbolt "github.com/ThreeDotsLabs/watermill-bolt/pkg/bolt"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/boreq/eggplant/adapters"
	"github.com/boreq/errors"
	bolt "go.etcd.io/bbolt"
)

type PubSub struct {
	subscriber *wmbolt.Subscriber
	router     *message.Router
}

func NewPubSub(db *bolt.DB) (*PubSub, error) {
	wmLogger := watermill.NopLogger{}

	subscriber, err := wmbolt.NewSubscriber(db, wmbolt.SubscriberConfig{Common: adapters.WatermillConfig()})
	if err != nil {
		return nil, errors.Wrap(err, "could not create the subscriber")
	}

	// Without an existing subscription bucket the publisher drops the message.
	for _, topic := range adapters.AllTopics() {
		if err := subscriber.SubscribeInitialize(topic); err != nil {
			return nil, errors.Wrap(err, "could not initialize the subscription")
		}
	}

	router, err := message.NewRouter(message.RouterConfig{}, wmLogger)
	if err != nil {
		return nil, errors.Wrap(err, "could not create the router")
	}

	return &PubSub{
		subscriber: subscriber,
		router:     router,
	}, nil
}

func (p *PubSub) AddHandler(name, topic string, fn message.NoPublishHandlerFunc) {
	p.router.AddConsumerHandler(name, topic, p.subscriber, fn)
}

func (p *PubSub) Run(ctx context.Context) error {
	return p.router.Run(ctx)
}
