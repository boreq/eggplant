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
	connect               *ConnectHandler
	disconnect            *DisconnectHandler
	setControllable       *SetControllableHandler
	rename                *RenameHandler
	takeControl           *TakeControlHandler
	releaseControl        *ReleaseControlHandler
	disconnectControllers *DisconnectControllersHandler
	sendCommand           *SendCommandHandler
	publishState          *PublishStateHandler
}

func NewControl(
	connect *ConnectHandler,
	disconnect *DisconnectHandler,
	setControllable *SetControllableHandler,
	rename *RenameHandler,
	takeControl *TakeControlHandler,
	releaseControl *ReleaseControlHandler,
	disconnectControllers *DisconnectControllersHandler,
	sendCommand *SendCommandHandler,
	publishState *PublishStateHandler,
) *Control {
	return &Control{
		connect:               connect,
		disconnect:            disconnect,
		setControllable:       setControllable,
		rename:                rename,
		takeControl:           takeControl,
		releaseControl:        releaseControl,
		disconnectControllers: disconnectControllers,
		sendCommand:           sendCommand,
		publishState:          publishState,
	}
}

func (c Control) Connect() *ConnectHandler {
	return c.connect
}

func (c Control) Disconnect() *DisconnectHandler {
	return c.disconnect
}

func (c Control) SetControllable() *SetControllableHandler {
	return c.setControllable
}

func (c Control) Rename() *RenameHandler {
	return c.rename
}

func (c Control) TakeControl() *TakeControlHandler {
	return c.takeControl
}

func (c Control) ReleaseControl() *ReleaseControlHandler {
	return c.releaseControl
}

func (c Control) DisconnectControllers() *DisconnectControllersHandler {
	return c.disconnectControllers
}

func (c Control) SendCommand() *SendCommandHandler {
	return c.sendCommand
}

func (c Control) PublishState() *PublishStateHandler {
	return c.publishState
}
