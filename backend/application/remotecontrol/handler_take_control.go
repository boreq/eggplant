package remotecontrol

import (
	"github.com/boreq/eggplant/domain/remotecontrol"
	"github.com/boreq/errors"
)

type TakeControl struct {
	connection remotecontrol.ConnectionID
	target     remotecontrol.DeviceID
}

func NewTakeControl(connection remotecontrol.ConnectionID, target remotecontrol.DeviceID) TakeControl {
	return TakeControl{
		connection: connection,
		target:     target,
	}
}

func (t TakeControl) Connection() remotecontrol.ConnectionID {
	return t.connection
}

func (t TakeControl) Target() remotecontrol.DeviceID {
	return t.target
}

type TakeControlHandler struct {
	hub Hub
}

func NewTakeControlHandler(hub Hub) *TakeControlHandler {
	return &TakeControlHandler{hub: hub}
}

func (h *TakeControlHandler) Execute(cmd TakeControl) error {
	if err := h.hub.TakeControl(cmd.Connection(), cmd.Target()); err != nil {
		return errors.Wrap(err, "could not take control of the device")
	}
	return nil
}
