package remotecontrol

import (
	"github.com/boreq/eggplant/domain/remotecontrol"
	"github.com/boreq/errors"
)

type SendCommand struct {
	connection remotecontrol.ConnectionID
	payload    remotecontrol.Payload
}

func NewSendCommand(connection remotecontrol.ConnectionID, payload remotecontrol.Payload) SendCommand {
	return SendCommand{
		connection: connection,
		payload:    payload,
	}
}

func (s SendCommand) Connection() remotecontrol.ConnectionID {
	return s.connection
}

func (s SendCommand) Payload() remotecontrol.Payload {
	return s.payload
}

type SendCommandHandler struct {
	hub Hub
}

func NewSendCommandHandler(hub Hub) *SendCommandHandler {
	return &SendCommandHandler{hub: hub}
}

func (h *SendCommandHandler) Execute(cmd SendCommand) error {
	if err := h.hub.SendCommand(cmd.Connection(), cmd.Payload()); err != nil {
		return errors.Wrap(err, "could not send the command")
	}
	return nil
}
