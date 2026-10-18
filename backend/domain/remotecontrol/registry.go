package remotecontrol

import (
	"cmp"
	"slices"

	"github.com/boreq/errors"
)

var (
	ErrConnectionNotFound = errors.New("connection not found")
	ErrConnectionExists   = errors.New("connection already exists")
	ErrDeviceNotFound     = errors.New("device not found")
	ErrCannotControlSelf  = errors.New("device cannot control itself")
	ErrDeviceIsTarget     = errors.New("device is a target")
	ErrTargetIsController = errors.New("target is a controller")
	ErrNotController      = errors.New("connection is not a controller")
	ErrNotControllable    = errors.New("connection is not controllable")
)

type Registry struct {
	connections map[ConnectionID]*connection
	devices     map[DeviceID]*deviceState
}

func NewRegistry() *Registry {
	return &Registry{
		connections: make(map[ConnectionID]*connection),
		devices:     make(map[DeviceID]*deviceState),
	}
}

type connection struct {
	device DeviceID
	target *DeviceID
}

func newConnection(device DeviceID) *connection {
	return &connection{device: device, target: nil}
}

func (c *connection) controls(target DeviceID) bool {
	return c.target != nil && *c.target == target
}

type deviceState struct {
	name         DeviceName
	controllable *ConnectionID
}

func newDeviceState(name DeviceName) *deviceState {
	return &deviceState{name: name, controllable: nil}
}

func (d *deviceState) isControllable(id ConnectionID) bool {
	return d.controllable != nil && *d.controllable == id
}

func (r *Registry) Empty() bool {
	return len(r.connections) == 0
}

func (r *Registry) Devices() []Device {
	return r.listDevices()
}

func (r *Registry) Connect(id ConnectionID, deviceID DeviceID, name DeviceName) ([]Delivery, error) {
	if _, ok := r.connections[id]; ok {
		return nil, ErrConnectionExists
	}

	r.connections[id] = newConnection(deviceID)

	d, ok := r.devices[deviceID]
	if !ok {
		d = newDeviceState(name)
		r.devices[deviceID] = d
	}
	d.name = name

	return r.broadcastDevices(), nil
}

func (r *Registry) Disconnect(id ConnectionID) []Delivery {
	c, ok := r.connections[id]
	if !ok {
		return nil
	}

	delete(r.connections, id)

	var deliveries []Delivery

	d := r.devices[c.device]
	if d.isControllable(id) {
		d.controllable = nil
		deliveries = append(deliveries, r.releaseControllers(c.device, ReleaseReasonOffline)...)
	}

	if !r.hasConnections(c.device) {
		delete(r.devices, c.device)
	}

	return append(deliveries, r.broadcastDevices()...)
}

func (r *Registry) SetControllable(id ConnectionID) ([]Delivery, error) {
	c, ok := r.connections[id]
	if !ok {
		return nil, ErrConnectionNotFound
	}

	r.devices[c.device].controllable = &id

	return r.broadcastDevices(), nil
}

func (r *Registry) Rename(id ConnectionID, name DeviceName) ([]Delivery, error) {
	c, ok := r.connections[id]
	if !ok {
		return nil, ErrConnectionNotFound
	}

	r.devices[c.device].name = name

	return r.broadcastDevices(), nil
}

func (r *Registry) TakeControl(id ConnectionID, target DeviceID) ([]Delivery, error) {
	c, ok := r.connections[id]
	if !ok {
		return nil, ErrConnectionNotFound
	}

	if target == c.device {
		return nil, ErrCannotControlSelf
	}

	t, ok := r.devices[target]
	if !ok || t.controllable == nil {
		return nil, ErrDeviceNotFound
	}

	if r.isTarget(c.device) {
		return nil, ErrDeviceIsTarget
	}

	if r.isController(target) {
		return nil, ErrTargetIsController
	}

	c.target = &target

	return r.broadcastDevices(), nil
}

func (r *Registry) ReleaseControl(id ConnectionID) ([]Delivery, error) {
	c, ok := r.connections[id]
	if !ok {
		return nil, ErrConnectionNotFound
	}

	if c.target == nil {
		return nil, nil
	}

	c.target = nil

	return r.broadcastDevices(), nil
}

func (r *Registry) DisconnectControllers(id ConnectionID) ([]Delivery, error) {
	c, ok := r.connections[id]
	if !ok {
		return nil, ErrConnectionNotFound
	}

	deliveries := r.releaseControllers(c.device, ReleaseReasonDisconnected)
	if len(deliveries) == 0 {
		return nil, nil
	}

	return append(deliveries, r.broadcastDevices()...), nil
}

func (r *Registry) Command(id ConnectionID, payload Payload) ([]Delivery, error) {
	c, ok := r.connections[id]
	if !ok {
		return nil, ErrConnectionNotFound
	}

	if c.target == nil {
		return nil, ErrNotController
	}

	return []Delivery{
		newDelivery(*r.devices[*c.target].controllable, newCommandEvent(payload)),
	}, nil
}

func (r *Registry) State(id ConnectionID, payload Payload) ([]Delivery, error) {
	c, ok := r.connections[id]
	if !ok {
		return nil, ErrConnectionNotFound
	}

	if !r.devices[c.device].isControllable(id) {
		return nil, ErrNotControllable
	}

	var deliveries []Delivery
	for controllerID, controller := range r.connections {
		if controller.controls(c.device) {
			deliveries = append(deliveries, newDelivery(controllerID, newStateEvent(payload)))
		}
	}
	return deliveries, nil
}

func (r *Registry) releaseControllers(target DeviceID, reason ReleaseReason) []Delivery {
	var deliveries []Delivery
	for id, c := range r.connections {
		if c.controls(target) {
			c.target = nil
			deliveries = append(deliveries, newDelivery(id, newReleasedEvent(reason)))
		}
	}
	return deliveries
}

func (r *Registry) broadcastDevices() []Delivery {
	event := newDevicesEvent(r.listDevices())

	deliveries := make([]Delivery, 0, len(r.connections))
	for id := range r.connections {
		deliveries = append(deliveries, newDelivery(id, event))
	}
	return deliveries
}

func (r *Registry) listDevices() []Device {
	devices := make([]Device, 0, len(r.devices))
	for id, d := range r.devices {
		if d.controllable == nil {
			continue
		}
		devices = append(devices, newDevice(id, d.name, r.mode(id), r.controllers(id)))
	}

	slices.SortFunc(devices, func(a, b Device) int {
		return cmp.Or(
			cmp.Compare(a.name.String(), b.name.String()),
			cmp.Compare(a.id.String(), b.id.String()),
		)
	})

	return devices
}

func (r *Registry) mode(id DeviceID) Mode {
	switch {
	case r.isTarget(id):
		return ModeTarget
	case r.isController(id):
		return ModeController
	default:
		return ModeLocal
	}
}

func (r *Registry) controllers(target DeviceID) []DeviceID {
	controllers := make([]DeviceID, 0)
	for _, c := range r.connections {
		if c.controls(target) && !slices.Contains(controllers, c.device) {
			controllers = append(controllers, c.device)
		}
	}

	slices.SortFunc(controllers, func(a, b DeviceID) int {
		return cmp.Compare(a.String(), b.String())
	})

	return controllers
}

func (r *Registry) isTarget(id DeviceID) bool {
	for _, c := range r.connections {
		if c.controls(id) {
			return true
		}
	}
	return false
}

func (r *Registry) isController(id DeviceID) bool {
	for _, c := range r.connections {
		if c.device == id && c.target != nil {
			return true
		}
	}
	return false
}

func (r *Registry) hasConnections(id DeviceID) bool {
	for _, c := range r.connections {
		if c.device == id {
			return true
		}
	}
	return false
}
