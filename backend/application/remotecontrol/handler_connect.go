package remotecontrol

import (
	"github.com/boreq/eggplant/application/accessctx"
	"github.com/boreq/eggplant/domain/remotecontrol"
	"github.com/boreq/errors"
)

type Connect struct {
	deviceID remotecontrol.DeviceID
	name     remotecontrol.DeviceName
}

func NewConnect(deviceID remotecontrol.DeviceID, name remotecontrol.DeviceName) Connect {
	return Connect{
		deviceID: deviceID,
		name:     name,
	}
}

func (c Connect) DeviceID() remotecontrol.DeviceID {
	return c.deviceID
}

func (c Connect) Name() remotecontrol.DeviceName {
	return c.name
}

type Connection struct {
	id     remotecontrol.ConnectionID
	events <-chan remotecontrol.Event
}

func NewConnection(id remotecontrol.ConnectionID, events <-chan remotecontrol.Event) Connection {
	return Connection{
		id:     id,
		events: events,
	}
}

func (c Connection) ID() remotecontrol.ConnectionID {
	return c.id
}

func (c Connection) Events() <-chan remotecontrol.Event {
	return c.events
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

	id, events, err := h.hub.Connect(userCtx.Username(), cmd.DeviceID(), cmd.Name())
	if err != nil {
		return Connection{}, errors.Wrap(err, "could not connect the device")
	}

	return NewConnection(id, events), nil
}
