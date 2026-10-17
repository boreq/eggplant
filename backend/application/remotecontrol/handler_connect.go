package remotecontrol

import (
	"github.com/boreq/eggplant/application/accessctx"
	"github.com/boreq/eggplant/domain/remotecontrol"
	"github.com/boreq/errors"
)

type Connect struct {
	DeviceID remotecontrol.DeviceID
	Name     remotecontrol.DeviceName
}

type Connection struct {
	ID     remotecontrol.ConnectionID
	Events <-chan remotecontrol.Event
}

type ConnectHandler struct {
	hub Hub
}

func NewConnectHandler(hub Hub) *ConnectHandler {
	return &ConnectHandler{hub: hub}
}

func (h *ConnectHandler) Execute(accessCtx accessctx.AccessContext, cmd Connect) (Connection, error) {
	userCtx, ok := accessCtx.(accessctx.UserAccessContext)
	if !ok {
		return Connection{}, accessctx.ErrPermissionDenied
	}

	id, events, err := h.hub.Connect(userCtx.Username(), cmd.DeviceID, cmd.Name)
	if err != nil {
		return Connection{}, errors.Wrap(err, "could not connect the device")
	}

	return Connection{ID: id, Events: events}, nil
}
