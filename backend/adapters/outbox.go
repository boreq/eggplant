package adapters

import (
	"fmt"

	wmbolt "github.com/ThreeDotsLabs/watermill-bolt/pkg/bolt"
	"github.com/ThreeDotsLabs/watermill/message"
	remoteadapter "github.com/boreq/eggplant/adapters/remote"
	remotedomain "github.com/boreq/eggplant/domain/remote"
	"github.com/boreq/errors"
	bolt "go.etcd.io/bbolt"
)

var TopicRemotePairingTokenSet = remotedomain.RemotePairingTokenSet{}.EventName()
var TopicRemotePaired = remotedomain.RemotePaired{}.EventName()

func AllTopics() []string {
	return []string{TopicRemotePairingTokenSet, TopicRemotePaired}
}

func WatermillConfig() wmbolt.CommonConfig {
	return wmbolt.CommonConfig{
		Bucket: []wmbolt.BucketName{wmbolt.BucketName("outbox")},
	}
}

type OutboxRepository struct {
	tx *bolt.Tx
}

func NewOutboxRepository(tx *bolt.Tx) (*OutboxRepository, error) {
	return &OutboxRepository{tx: tx}, nil
}

func (r *OutboxRepository) AddEvents(source remotedomain.EventSource) error {
	publisher, err := wmbolt.NewTxPublisher(r.tx, wmbolt.PublisherConfig{Common: WatermillConfig()})
	if err != nil {
		return errors.Wrap(err, "could not create the tx publisher")
	}

	for _, event := range source.PopEvents() {
		topic, msg, err := marshalEvent(event)
		if err != nil {
			return errors.Wrap(err, "could not marshal the event")
		}

		if err := publisher.Publish(topic, msg); err != nil {
			return errors.Wrap(err, "could not publish the event")
		}
	}

	return nil
}

func marshalEvent(event remotedomain.Event) (string, *message.Message, error) {
	switch e := event.(type) {
	case remotedomain.RemotePairingTokenSet:
		msg, err := remoteadapter.MarshalRemotePairingTokenSet(e)
		if err != nil {
			return "", nil, errors.Wrap(err, "could not marshal the event")
		}
		return e.EventName(), msg, nil
	case remotedomain.RemotePaired:
		msg, err := remoteadapter.MarshalRemotePaired(e)
		if err != nil {
			return "", nil, errors.Wrap(err, "could not marshal the event")
		}
		return e.EventName(), msg, nil
	default:
		return "", nil, errors.New(fmt.Sprintf("unknown event type %T", event))
	}
}
