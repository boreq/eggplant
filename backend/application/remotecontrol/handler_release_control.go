package remotecontrol

import (
	"github.com/boreq/eggplant/domain/remotecontrol"
	"github.com/boreq/errors"
)

type ReleaseControl struct {
	Connection remotecontrol.ConnectionID
}

type ReleaseControlHandler struct {
	hub Hub
}

func NewReleaseControlHandler(hub Hub) *ReleaseControlHandler {
	return &ReleaseControlHandler{hub: hub}
}

func (h *ReleaseControlHandler) Execute(cmd ReleaseControl) error {
	if err := h.hub.ReleaseControl(cmd.Connection); err != nil {
		return errors.Wrap(err, "could not release control")
	}
	return nil
}
