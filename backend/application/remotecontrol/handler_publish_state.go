package remotecontrol

import (
	"github.com/boreq/eggplant/domain/remotecontrol"
	"github.com/boreq/errors"
)

type PublishState struct {
	Connection remotecontrol.ConnectionID
	Payload    remotecontrol.Payload
}

type PublishStateHandler struct {
	hub Hub
}

func NewPublishStateHandler(hub Hub) *PublishStateHandler {
	return &PublishStateHandler{hub: hub}
}

func (h *PublishStateHandler) Execute(cmd PublishState) error {
	if err := h.hub.PublishState(cmd.Connection, cmd.Payload); err != nil {
		return errors.Wrap(err, "could not publish the state")
	}
	return nil
}
