package remotecontrol

import (
	"github.com/boreq/eggplant/domain/remotecontrol"
	"github.com/boreq/errors"
)

type PublishState struct {
	connection remotecontrol.ConnectionID
	payload    remotecontrol.Payload
}

func NewPublishState(connection remotecontrol.ConnectionID, payload remotecontrol.Payload) PublishState {
	return PublishState{
		connection: connection,
		payload:    payload,
	}
}

func (p PublishState) Connection() remotecontrol.ConnectionID {
	return p.connection
}

func (p PublishState) Payload() remotecontrol.Payload {
	return p.payload
}

type PublishStateHandler struct {
	hub Hub
}

func NewPublishStateHandler(hub Hub) *PublishStateHandler {
	return &PublishStateHandler{hub: hub}
}

func (h *PublishStateHandler) Execute(cmd PublishState) error {
	if err := h.hub.PublishState(cmd.Connection(), cmd.Payload()); err != nil {
		return errors.Wrap(err, "could not publish the state")
	}
	return nil
}
