package outbox

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/boreq/eggplant/adapters"
)

type Subscriber interface {
	AddHandler(name, topic string, fn message.NoPublishHandlerFunc)
	Run(ctx context.Context) error
}

type Listener struct {
	subscriber Subscriber
}

func NewListener(subscriber Subscriber, sendLocalAuthToken SendLocalAuthTokenHandler, checkRemote CheckRemoteHandler) *Listener {
	l := &Listener{subscriber: subscriber}

	pairingTokenSet := NewRemotePairingTokenSetHandler(sendLocalAuthToken)
	paired := NewRemotePairedHandler(checkRemote)

	l.subscriber.AddHandler("remote_pairing_token_set", adapters.TopicRemotePairingTokenSet, pairingTokenSet.Handle)
	l.subscriber.AddHandler("remote_paired", adapters.TopicRemotePaired, paired.Handle)

	return l
}

func (l *Listener) Run(ctx context.Context) error {
	return l.subscriber.Run(ctx)
}
