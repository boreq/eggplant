package remotecontrol

import (
	"github.com/boreq/eggplant/domain/remotecontrol"
	"github.com/boreq/errors"
)

type SendCommand struct {
	Connection remotecontrol.ConnectionID
	Payload    remotecontrol.Payload
}

type SendCommandHandler struct {
	hub Hub
}

func NewSendCommandHandler(hub Hub) *SendCommandHandler {
	return &SendCommandHandler{hub: hub}
}

func (h *SendCommandHandler) Execute(cmd SendCommand) error {
	if err := h.hub.SendCommand(cmd.Connection, cmd.Payload); err != nil {
		return errors.Wrap(err, "could not send the command")
	}
	return nil
}
