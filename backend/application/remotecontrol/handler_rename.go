package remotecontrol

import (
	"github.com/boreq/eggplant/domain/remotecontrol"
	"github.com/boreq/errors"
)

type Rename struct {
	Connection remotecontrol.ConnectionID
	Name       remotecontrol.DeviceName
}

type RenameHandler struct {
	hub Hub
}

func NewRenameHandler(hub Hub) *RenameHandler {
	return &RenameHandler{hub: hub}
}

func (h *RenameHandler) Execute(cmd Rename) error {
	if err := h.hub.Rename(cmd.Connection, cmd.Name); err != nil {
		return errors.Wrap(err, "could not rename the device")
	}
	return nil
}
