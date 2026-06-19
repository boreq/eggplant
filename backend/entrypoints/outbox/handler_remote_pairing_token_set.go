package outbox

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/boreq/eggplant/adapters"
	remoteadapter "github.com/boreq/eggplant/adapters/remote"
	remoteapp "github.com/boreq/eggplant/application/remote"
	"github.com/boreq/eggplant/internal/logging"
	"github.com/boreq/errors"
)

type SendLocalAuthTokenHandler interface {
	Execute(ctx context.Context, cmd remoteapp.SendLocalAuthToken) error
}

type RemotePairingTokenSetHandler struct {
	sendLocalAuthToken SendLocalAuthTokenHandler
	log                logging.Logger
}

func NewRemotePairingTokenSetHandler(sendLocalAuthToken SendLocalAuthTokenHandler) *RemotePairingTokenSetHandler {
	return &RemotePairingTokenSetHandler{
		sendLocalAuthToken: sendLocalAuthToken,
		log:                logging.New("entrypoints/outbox.RemotePairingTokenSetHandler"),
	}
}

func (h *RemotePairingTokenSetHandler) Handle(msg *message.Message) error {
	event, err := remoteadapter.UnmarshalRemotePairingTokenSet(msg)
	if err != nil {
		h.log.Error("could not unmarshal an event", "topic", adapters.TopicRemotePairingTokenSet, "err", err)
		return errors.Wrap(err, "could not unmarshal the event")
	}

	if err := h.sendLocalAuthToken.Execute(msg.Context(), remoteapp.SendLocalAuthToken{RemoteInstanceID: event.RemoteInstanceID}); err != nil {
		h.log.Error("could not handle an event", "topic", adapters.TopicRemotePairingTokenSet, "err", err)
		return err
	}

	return nil
}
