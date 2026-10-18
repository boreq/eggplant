package remotecontrol

import (
	"github.com/boreq/eggplant/domain/remotecontrol"
	"github.com/boreq/errors"
)

type DisconnectControllers struct {
	Connection remotecontrol.ConnectionID
}

type DisconnectControllersHandler struct {
	hub Hub
}

func NewDisconnectControllersHandler(hub Hub) *DisconnectControllersHandler {
	return &DisconnectControllersHandler{hub: hub}
}

func (h *DisconnectControllersHandler) Execute(cmd DisconnectControllers) error {
	if err := h.hub.DisconnectControllers(cmd.Connection); err != nil {
		return errors.Wrap(err, "could not disconnect the controllers")
	}
	return nil
}
