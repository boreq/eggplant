package remotecontrol_test

import (
	"testing"

	remotecontroladapter "github.com/boreq/eggplant/adapters/remotecontrol"
	authdomain "github.com/boreq/eggplant/domain/auth"
	"github.com/boreq/eggplant/domain/remotecontrol"
	"github.com/stretchr/testify/require"
)

func TestHub_UsersAreIsolated(t *testing.T) {
	hub := remotecontroladapter.NewHub()

	alice, aliceEvents := connect(t, hub, "alice", "a")
	require.NoError(t, hub.SetControllable(alice))
	_, bobEvents := connect(t, hub, "bob", "b")

	drain(aliceEvents)
	devices := lastDevicesEvent(t, bobEvents)
	require.Empty(t, devices.Devices())

	bob2, _ := connect(t, hub, "bob", "b2")
	require.ErrorIs(t, hub.TakeControl(bob2, deviceID(t, "a")), remotecontrol.ErrDeviceNotFound)
}

func TestHub_CommandAndStateAreDelivered(t *testing.T) {
	hub := remotecontroladapter.NewHub()

	a, aEvents := connect(t, hub, "alice", "a")
	require.NoError(t, hub.SetControllable(a))
	b, bEvents := connect(t, hub, "alice", "b")
	require.NoError(t, hub.SetControllable(b))

	require.NoError(t, hub.TakeControl(a, deviceID(t, "b")))
	drain(aEvents)
	drain(bEvents)

	require.NoError(t, hub.SendCommand(a, payload(t, "next")))
	command, ok := (<-bEvents).(remotecontrol.CommandEvent)
	require.True(t, ok)
	require.Equal(t, payload(t, "next"), command.Payload())

	require.NoError(t, hub.PublishState(b, payload(t, "state")))
	state, ok := (<-aEvents).(remotecontrol.StateEvent)
	require.True(t, ok)
	require.Equal(t, payload(t, "state"), state.Payload())
}

func TestHub_DisconnectClosesEventsAndReleasesControllers(t *testing.T) {
	hub := remotecontroladapter.NewHub()

	a, aEvents := connect(t, hub, "alice", "a")
	require.NoError(t, hub.SetControllable(a))
	b, bEvents := connect(t, hub, "alice", "b")
	require.NoError(t, hub.SetControllable(b))

	require.NoError(t, hub.TakeControl(a, deviceID(t, "b")))
	drain(aEvents)

	hub.Disconnect(b)

	for range bEvents {
	}

	released, ok := (<-aEvents).(remotecontrol.ReleasedEvent)
	require.True(t, ok)
	require.Equal(t, remotecontrol.ReleaseReasonOffline, released.Reason())

	require.ErrorIs(t, hub.SetControllable(b), remotecontrol.ErrConnectionNotFound)
}

func TestHub_SlowConnectionIsDropped(t *testing.T) {
	hub := remotecontroladapter.NewHub()

	a, aEvents := connect(t, hub, "alice", "a")
	require.NoError(t, hub.SetControllable(a))
	b, _ := connect(t, hub, "alice", "b")

	for range 100 {
		require.NoError(t, hub.Rename(a, deviceName(t, "a")))
		drain(aEvents)
	}

	require.ErrorIs(t, hub.SetControllable(b), remotecontrol.ErrConnectionNotFound)
}

func connect(t *testing.T, hub *remotecontroladapter.Hub, username, id string) (remotecontrol.ConnectionID, <-chan remotecontrol.Event) {
	u, err := authdomain.NewUsernameFromString(username)
	require.NoError(t, err)

	connectionID, events, err := hub.Connect(u, deviceID(t, id), deviceName(t, id))
	require.NoError(t, err)

	return connectionID, events
}

func drain(events <-chan remotecontrol.Event) {
	for {
		select {
		case <-events:
		default:
			return
		}
	}
}

func lastDevicesEvent(t *testing.T, events <-chan remotecontrol.Event) remotecontrol.DevicesEvent {
	var last remotecontrol.DevicesEvent
	found := false
	for {
		select {
		case event := <-events:
			if e, ok := event.(remotecontrol.DevicesEvent); ok {
				last = e
				found = true
			}
		default:
			require.True(t, found)
			return last
		}
	}
}

func deviceID(t *testing.T, s string) remotecontrol.DeviceID {
	id, err := remotecontrol.NewDeviceID(s)
	require.NoError(t, err)
	return id
}

func deviceName(t *testing.T, s string) remotecontrol.DeviceName {
	name, err := remotecontrol.NewDeviceName(s)
	require.NoError(t, err)
	return name
}

func payload(t *testing.T, s string) remotecontrol.Payload {
	p, err := remotecontrol.NewPayload([]byte(s))
	require.NoError(t, err)
	return p
}
