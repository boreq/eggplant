package remotecontrol

import (
	"github.com/boreq/eggplant/domain/remotecontrol"
	"github.com/boreq/errors"
)

type SetControllable struct {
	Connection remotecontrol.ConnectionID
}

type SetControllableHandler struct {
	hub Hub
}

func NewSetControllableHandler(hub Hub) *SetControllableHandler {
	return &SetControllableHandler{hub: hub}
}

func (h *SetControllableHandler) Execute(cmd SetControllable) error {
	if err := h.hub.SetControllable(cmd.Connection); err != nil {
		return errors.Wrap(err, "could not mark the connection as controllable")
	}
	return nil
}
