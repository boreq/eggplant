package remotecontrol

import (
	"github.com/boreq/eggplant/domain/remotecontrol"
	"github.com/boreq/errors"
)

type TakeControl struct {
	Connection remotecontrol.ConnectionID
	Target     remotecontrol.DeviceID
}

type TakeControlHandler struct {
	hub Hub
}

func NewTakeControlHandler(hub Hub) *TakeControlHandler {
	return &TakeControlHandler{hub: hub}
}

func (h *TakeControlHandler) Execute(cmd TakeControl) error {
	if err := h.hub.TakeControl(cmd.Connection, cmd.Target); err != nil {
		return errors.Wrap(err, "could not take control of the device")
	}
	return nil
}
