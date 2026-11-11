package remotecontrol

import (
	"sync"

	authdomain "github.com/boreq/eggplant/domain/auth"
	"github.com/boreq/eggplant/domain/remotecontrol"
	"github.com/boreq/eggplant/internal/logging"
	"github.com/boreq/errors"
)

const eventBufferSize = 64

type Hub struct {
	mutex       sync.Mutex
	registries  map[authdomain.Username]*remotecontrol.Registry
	connections map[remotecontrol.ConnectionID]*hubConnection
	log         logging.Logger
}

type hubConnection struct {
	username authdomain.Username
	events   chan remotecontrol.Event
}

func NewHub() *Hub {
	return &Hub{
		registries:  make(map[authdomain.Username]*remotecontrol.Registry),
		connections: make(map[remotecontrol.ConnectionID]*hubConnection),
		log:         logging.New("adapters/remotecontrol.Hub"),
	}
}

func (h *Hub) Connect(username authdomain.Username, deviceID remotecontrol.DeviceID, name remotecontrol.DeviceName) (remotecontrol.ConnectionID, <-chan remotecontrol.Event, error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	id := remotecontrol.NewConnectionID()

	registry, ok := h.registries[username]
	if !ok {
		registry = remotecontrol.NewRegistry()
		h.registries[username] = registry
	}

	events := make(chan remotecontrol.Event, eventBufferSize)
	h.connections[id] = &hubConnection{
		username: username,
		events:   events,
	}

	deliveries, err := registry.Connect(id, deviceID, name)
	if err != nil {
		delete(h.connections, id)
		if registry.Empty() {
			delete(h.registries, username)
		}
		return remotecontrol.ConnectionID{}, nil, errors.Wrap(err, "could not connect")
	}

	h.deliver(deliveries)

	return id, events, nil
}

func (h *Hub) Disconnect(id remotecontrol.ConnectionID) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.disconnect(id)
}

func (h *Hub) SetControllable(id remotecontrol.ConnectionID) error {
	return h.update(id, func(r *remotecontrol.Registry) ([]remotecontrol.Delivery, error) {
		return r.SetControllable(id)
	})
}

func (h *Hub) Rename(id remotecontrol.ConnectionID, name remotecontrol.DeviceName) error {
	return h.update(id, func(r *remotecontrol.Registry) ([]remotecontrol.Delivery, error) {
		return r.Rename(id, name)
	})
}

func (h *Hub) TakeControl(id remotecontrol.ConnectionID, target remotecontrol.DeviceID) error {
	return h.update(id, func(r *remotecontrol.Registry) ([]remotecontrol.Delivery, error) {
		return r.TakeControl(id, target)
	})
}

func (h *Hub) ReleaseControl(id remotecontrol.ConnectionID) error {
	return h.update(id, func(r *remotecontrol.Registry) ([]remotecontrol.Delivery, error) {
		return r.ReleaseControl(id)
	})
}

func (h *Hub) DisconnectControllers(id remotecontrol.ConnectionID) error {
	return h.update(id, func(r *remotecontrol.Registry) ([]remotecontrol.Delivery, error) {
		return r.DisconnectControllers(id)
	})
}

func (h *Hub) SendCommand(id remotecontrol.ConnectionID, payload remotecontrol.Payload) error {
	return h.update(id, func(r *remotecontrol.Registry) ([]remotecontrol.Delivery, error) {
		return r.Command(id, payload)
	})
}

func (h *Hub) PublishState(id remotecontrol.ConnectionID, payload remotecontrol.Payload) error {
	return h.update(id, func(r *remotecontrol.Registry) ([]remotecontrol.Delivery, error) {
		return r.State(id, payload)
	})
}

func (h *Hub) update(id remotecontrol.ConnectionID, fn func(r *remotecontrol.Registry) ([]remotecontrol.Delivery, error)) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	c, ok := h.connections[id]
	if !ok {
		return remotecontrol.ErrConnectionNotFound
	}

	deliveries, err := fn(h.registries[c.username])
	if err != nil {
		return err
	}

	h.deliver(deliveries)

	return nil
}

func (h *Hub) disconnect(id remotecontrol.ConnectionID) {
	c, ok := h.connections[id]
	if !ok {
		return
	}

	delete(h.connections, id)
	close(c.events)

	registry := h.registries[c.username]
	deliveries := registry.Disconnect(id)
	if registry.Empty() {
		delete(h.registries, c.username)
	}

	h.deliver(deliveries)
}

func (h *Hub) deliver(deliveries []remotecontrol.Delivery) {
	for _, d := range deliveries {
		c, ok := h.connections[d.Connection()]
		if !ok {
			continue
		}

		select {
		case c.events <- d.Event():
		default:
			h.log.Warn("dropping a slow connection", "connection", d.Connection().String())
			h.disconnect(d.Connection())
		}
	}
}
