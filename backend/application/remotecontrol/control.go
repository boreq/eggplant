package remotecontrol

import (
	authdomain "github.com/boreq/eggplant/domain/auth"
	"github.com/boreq/eggplant/domain/remotecontrol"
)

type Hub interface {
	Connect(username authdomain.Username, deviceID remotecontrol.DeviceID, name remotecontrol.DeviceName) (remotecontrol.ConnectionID, <-chan remotecontrol.Event, error)
	Disconnect(id remotecontrol.ConnectionID)
	SetControllable(id remotecontrol.ConnectionID) error
	Rename(id remotecontrol.ConnectionID, name remotecontrol.DeviceName) error
	TakeControl(id remotecontrol.ConnectionID, target remotecontrol.DeviceID) error
	ReleaseControl(id remotecontrol.ConnectionID) error
	DisconnectControllers(id remotecontrol.ConnectionID) error
	SendCommand(id remotecontrol.ConnectionID, payload remotecontrol.Payload) error
	PublishState(id remotecontrol.ConnectionID, payload remotecontrol.Payload) error
}

type Control struct {
	Connect               *ConnectHandler
	Disconnect            *DisconnectHandler
	SetControllable       *SetControllableHandler
	Rename                *RenameHandler
	TakeControl           *TakeControlHandler
	ReleaseControl        *ReleaseControlHandler
	DisconnectControllers *DisconnectControllersHandler
	SendCommand           *SendCommandHandler
	PublishState          *PublishStateHandler
}
