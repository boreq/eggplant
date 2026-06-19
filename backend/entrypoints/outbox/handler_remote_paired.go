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

type CheckRemoteHandler interface {
	Execute(ctx context.Context, cmd remoteapp.CheckRemote) error
}

type RemotePairedHandler struct {
	checkRemote CheckRemoteHandler
	log         logging.Logger
}

func NewRemotePairedHandler(checkRemote CheckRemoteHandler) *RemotePairedHandler {
	return &RemotePairedHandler{
		checkRemote: checkRemote,
		log:         logging.New("entrypoints/outbox.RemotePairedHandler"),
	}
}

func (h *RemotePairedHandler) Handle(msg *message.Message) error {
	event, err := remoteadapter.UnmarshalRemotePaired(msg)
	if err != nil {
		h.log.Error("could not unmarshal an event", "topic", adapters.TopicRemotePaired, "err", err)
		return errors.Wrap(err, "could not unmarshal the event")
	}

	if err := h.checkRemote.Execute(msg.Context(), remoteapp.CheckRemote{ID: event.RemoteInstanceID}); err != nil {
		h.log.Error("could not handle an event", "topic", adapters.TopicRemotePaired, "err", err)
		return err
	}

	return nil
}
