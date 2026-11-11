package remotecontrol

import (
	"github.com/boreq/eggplant/domain/remotecontrol"
	"github.com/boreq/errors"
)

type Rename struct {
	connection remotecontrol.ConnectionID
	name       remotecontrol.DeviceName
}

func NewRename(connection remotecontrol.ConnectionID, name remotecontrol.DeviceName) Rename {
	return Rename{
		connection: connection,
		name:       name,
	}
}

func (r Rename) Connection() remotecontrol.ConnectionID {
	return r.connection
}

func (r Rename) Name() remotecontrol.DeviceName {
	return r.name
}

type RenameHandler struct {
	hub Hub
}

func NewRenameHandler(hub Hub) *RenameHandler {
	return &RenameHandler{hub: hub}
}

func (h *RenameHandler) Execute(cmd Rename) error {
	if err := h.hub.Rename(cmd.Connection(), cmd.Name()); err != nil {
		return errors.Wrap(err, "could not rename the device")
	}
	return nil
}
