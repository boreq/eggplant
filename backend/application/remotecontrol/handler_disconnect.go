package remotecontrol

import "github.com/boreq/eggplant/domain/remotecontrol"

type Disconnect struct {
	Connection remotecontrol.ConnectionID
}

type DisconnectHandler struct {
	hub Hub
}

func NewDisconnectHandler(hub Hub) *DisconnectHandler {
	return &DisconnectHandler{hub: hub}
}

func (h *DisconnectHandler) Execute(cmd Disconnect) {
	h.hub.Disconnect(cmd.Connection)
}
