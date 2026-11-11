package remotecontrol

import (
	"github.com/boreq/eggplant/domain/remotecontrol"
	"github.com/boreq/errors"
)

type ReleaseControl struct {
	connection remotecontrol.ConnectionID
}

func NewReleaseControl(connection remotecontrol.ConnectionID) ReleaseControl {
	return ReleaseControl{
		connection: connection,
	}
}

func (r ReleaseControl) Connection() remotecontrol.ConnectionID {
	return r.connection
}

type ReleaseControlHandler struct {
	hub Hub
}

func NewReleaseControlHandler(hub Hub) *ReleaseControlHandler {
	return &ReleaseControlHandler{hub: hub}
}

func (h *ReleaseControlHandler) Execute(cmd ReleaseControl) error {
	if err := h.hub.ReleaseControl(cmd.Connection()); err != nil {
		return errors.Wrap(err, "could not release control")
	}
	return nil
}
