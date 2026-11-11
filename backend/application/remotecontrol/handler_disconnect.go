package remotecontrol

import "github.com/boreq/eggplant/domain/remotecontrol"

type Disconnect struct {
	connection remotecontrol.ConnectionID
}

func NewDisconnect(connection remotecontrol.ConnectionID) Disconnect {
	return Disconnect{
		connection: connection,
	}
}

func (d Disconnect) Connection() remotecontrol.ConnectionID {
	return d.connection
}

type DisconnectHandler struct {
	hub Hub
}

func NewDisconnectHandler(hub Hub) *DisconnectHandler {
	return &DisconnectHandler{hub: hub}
}

func (h *DisconnectHandler) Execute(cmd Disconnect) {
	h.hub.Disconnect(cmd.Connection())
}
