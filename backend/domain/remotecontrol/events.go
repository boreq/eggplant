package remotecontrol

import "slices"

type Mode struct {
	value string
}

var (
	ModeLocal      = Mode{value: "local"}
	ModeController = Mode{value: "controller"}
	ModeTarget     = Mode{value: "target"}
)

func (m Mode) String() string {
	return m.value
}

type ReleaseReason struct {
	value string
}

var (
	ReleaseReasonOffline      = ReleaseReason{value: "offline"}
	ReleaseReasonDisconnected = ReleaseReason{value: "disconnected"}
)

func (r ReleaseReason) String() string {
	return r.value
}

type Device struct {
	id          DeviceID
	name        DeviceName
	mode        Mode
	controllers []DeviceID
}

func newDevice(id DeviceID, name DeviceName, mode Mode, controllers []DeviceID) Device {
	return Device{id: id, name: name, mode: mode, controllers: controllers}
}

func (d Device) ID() DeviceID {
	return d.id
}

func (d Device) Name() DeviceName {
	return d.name
}

func (d Device) Mode() Mode {
	return d.mode
}

func (d Device) Controllers() []DeviceID {
	return slices.Clone(d.controllers)
}

type Event interface {
	isEvent()
}

type DevicesEvent struct {
	devices []Device
}

func newDevicesEvent(devices []Device) DevicesEvent {
	return DevicesEvent{devices: devices}
}

func (e DevicesEvent) Devices() []Device {
	return slices.Clone(e.devices)
}

type CommandEvent struct {
	payload Payload
}

func newCommandEvent(payload Payload) CommandEvent {
	return CommandEvent{payload: payload}
}

func (e CommandEvent) Payload() Payload {
	return e.payload
}

type StateEvent struct {
	payload Payload
}

func newStateEvent(payload Payload) StateEvent {
	return StateEvent{payload: payload}
}

func (e StateEvent) Payload() Payload {
	return e.payload
}

type ReleasedEvent struct {
	reason ReleaseReason
}

func newReleasedEvent(reason ReleaseReason) ReleasedEvent {
	return ReleasedEvent{reason: reason}
}

func (e ReleasedEvent) Reason() ReleaseReason {
	return e.reason
}

func (DevicesEvent) isEvent()  {}
func (CommandEvent) isEvent()  {}
func (StateEvent) isEvent()    {}
func (ReleasedEvent) isEvent() {}

type Delivery struct {
	connection ConnectionID
	event      Event
}

func newDelivery(connection ConnectionID, event Event) Delivery {
	return Delivery{connection: connection, event: event}
}

func (d Delivery) Connection() ConnectionID {
	return d.connection
}

func (d Delivery) Event() Event {
	return d.event
}
