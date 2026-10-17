package remotecontrol_test

import (
	"testing"

	"github.com/boreq/eggplant/domain/remotecontrol"
	"github.com/stretchr/testify/require"
)

func TestRegistry_ConnectBroadcastsDevicesToAllConnections(t *testing.T) {
	r := remotecontrol.NewRegistry()

	a := connect(t, r, "a", "Device A")
	setControllable(t, r, a)
	b := connect(t, r, "b", "Device B")

	deliveries, err := r.SetControllable(b)
	require.NoError(t, err)

	var connections []remotecontrol.ConnectionID
	for _, d := range deliveries {
		connections = append(connections, d.Connection())

		event, ok := d.Event().(remotecontrol.DevicesEvent)
		require.True(t, ok)

		devices := event.Devices()
		require.Len(t, devices, 2)
		require.Equal(t, deviceID(t, "a"), devices[0].ID())
		require.Equal(t, deviceName(t, "Device A"), devices[0].Name())
		require.Equal(t, deviceID(t, "b"), devices[1].ID())
		require.Equal(t, deviceName(t, "Device B"), devices[1].Name())
		for _, device := range devices {
			require.Equal(t, remotecontrol.ModeLocal, device.Mode())
			require.Empty(t, device.Controllers())
		}
	}
	require.ElementsMatch(t, []remotecontrol.ConnectionID{a, b}, connections)
}

func TestRegistry_DevicesWithoutControllableConnectionAreNotListed(t *testing.T) {
	r := remotecontrol.NewRegistry()

	a := connect(t, r, "a", "Device A")
	setControllable(t, r, a)
	connect(t, r, "b", "Device B")

	require.Equal(t, []remotecontrol.DeviceID{deviceID(t, "a")}, listedDevices(r))
}

func TestRegistry_TakeControl(t *testing.T) {
	r := remotecontrol.NewRegistry()

	a := connect(t, r, "a", "Device A")
	setControllable(t, r, a)
	b := connect(t, r, "b", "Device B")
	setControllable(t, r, b)

	_, err := r.TakeControl(a, deviceID(t, "b"))
	require.NoError(t, err)

	devices := lastDevices(r)
	require.Equal(t, remotecontrol.ModeController, devices[deviceID(t, "a")].Mode())
	require.Equal(t, remotecontrol.ModeTarget, devices[deviceID(t, "b")].Mode())
	require.Equal(t, []remotecontrol.DeviceID{deviceID(t, "a")}, devices[deviceID(t, "b")].Controllers())
}

func TestRegistry_TakeControlRules(t *testing.T) {
	testCases := []struct {
		Name          string
		Setup         func(t *testing.T, r *remotecontrol.Registry, conns map[string]remotecontrol.ConnectionID)
		Connection    string
		Target        string
		ExpectedError error
	}{
		{
			Name:          "device_cannot_control_itself",
			Connection:    "a",
			Target:        "a",
			ExpectedError: remotecontrol.ErrCannotControlSelf,
		},
		{
			Name:          "unknown_device",
			Connection:    "a",
			Target:        "unknown",
			ExpectedError: remotecontrol.ErrDeviceNotFound,
		},
		{
			Name: "device_without_controllable_connection",
			Setup: func(t *testing.T, r *remotecontrol.Registry, conns map[string]remotecontrol.ConnectionID) {
				conns["d"] = connect(t, r, "d", "Device D")
			},
			Connection:    "a",
			Target:        "d",
			ExpectedError: remotecontrol.ErrDeviceNotFound,
		},
		{
			Name: "controller_cannot_become_target",
			Setup: func(t *testing.T, r *remotecontrol.Registry, conns map[string]remotecontrol.ConnectionID) {
				takeControl(t, r, conns["b"], "c")
			},
			Connection:    "a",
			Target:        "b",
			ExpectedError: remotecontrol.ErrTargetIsController,
		},
		{
			Name: "target_cannot_become_controller",
			Setup: func(t *testing.T, r *remotecontrol.Registry, conns map[string]remotecontrol.ConnectionID) {
				takeControl(t, r, conns["b"], "a")
			},
			Connection:    "a",
			Target:        "c",
			ExpectedError: remotecontrol.ErrDeviceIsTarget,
		},
		{
			Name: "second_selection_fails_when_devices_select_each_other",
			Setup: func(t *testing.T, r *remotecontrol.Registry, conns map[string]remotecontrol.ConnectionID) {
				takeControl(t, r, conns["a"], "b")
			},
			Connection:    "b",
			Target:        "a",
			ExpectedError: remotecontrol.ErrDeviceIsTarget,
		},
		{
			Name: "other_tab_of_target_cannot_become_controller",
			Setup: func(t *testing.T, r *remotecontrol.Registry, conns map[string]remotecontrol.ConnectionID) {
				conns["a2"] = connect(t, r, "a", "Device A")
				takeControl(t, r, conns["b"], "a")
			},
			Connection:    "a2",
			Target:        "c",
			ExpectedError: remotecontrol.ErrDeviceIsTarget,
		},
		{
			Name: "device_with_controlling_tab_cannot_become_target",
			Setup: func(t *testing.T, r *remotecontrol.Registry, conns map[string]remotecontrol.ConnectionID) {
				conns["b2"] = connect(t, r, "b", "Device B")
				takeControl(t, r, conns["b2"], "c")
			},
			Connection:    "a",
			Target:        "b",
			ExpectedError: remotecontrol.ErrTargetIsController,
		},
		{
			Name: "several_controllers_can_control_one_target",
			Setup: func(t *testing.T, r *remotecontrol.Registry, conns map[string]remotecontrol.ConnectionID) {
				takeControl(t, r, conns["b"], "c")
			},
			Connection: "a",
			Target:     "c",
		},
		{
			Name: "controller_can_switch_targets",
			Setup: func(t *testing.T, r *remotecontrol.Registry, conns map[string]remotecontrol.ConnectionID) {
				takeControl(t, r, conns["a"], "b")
			},
			Connection: "a",
			Target:     "c",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			r := remotecontrol.NewRegistry()

			conns := make(map[string]remotecontrol.ConnectionID)
			for _, name := range []string{"a", "b", "c"} {
				conns[name] = connect(t, r, name, "Device "+name)
				setControllable(t, r, conns[name])
			}

			if testCase.Setup != nil {
				testCase.Setup(t, r, conns)
			}

			_, err := r.TakeControl(conns[testCase.Connection], deviceID(t, testCase.Target))
			if testCase.ExpectedError != nil {
				require.ErrorIs(t, err, testCase.ExpectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRegistry_ControllerSwitchingTargetsLeavesPreviousTarget(t *testing.T) {
	r := remotecontrol.NewRegistry()

	a := connect(t, r, "a", "Device A")
	setControllable(t, r, a)
	b := connect(t, r, "b", "Device B")
	setControllable(t, r, b)
	c := connect(t, r, "c", "Device C")
	setControllable(t, r, c)

	takeControl(t, r, a, "b")
	takeControl(t, r, a, "c")

	devices := lastDevices(r)
	require.Equal(t, remotecontrol.ModeLocal, devices[deviceID(t, "b")].Mode())
	require.Equal(t, remotecontrol.ModeTarget, devices[deviceID(t, "c")].Mode())
}

func TestRegistry_ReleaseControl(t *testing.T) {
	r := remotecontrol.NewRegistry()

	a := connect(t, r, "a", "Device A")
	setControllable(t, r, a)
	b := connect(t, r, "b", "Device B")
	setControllable(t, r, b)

	takeControl(t, r, a, "b")

	_, err := r.ReleaseControl(a)
	require.NoError(t, err)

	devices := lastDevices(r)
	require.Equal(t, remotecontrol.ModeLocal, devices[deviceID(t, "a")].Mode())
	require.Equal(t, remotecontrol.ModeLocal, devices[deviceID(t, "b")].Mode())

	_, err = r.Command(a, payload(t, "next"))
	require.ErrorIs(t, err, remotecontrol.ErrNotController)
}

func TestRegistry_TargetStaysTargetUntilLastControllerLeaves(t *testing.T) {
	r := remotecontrol.NewRegistry()

	a := connect(t, r, "a", "Device A")
	setControllable(t, r, a)
	b := connect(t, r, "b", "Device B")
	setControllable(t, r, b)
	c := connect(t, r, "c", "Device C")
	setControllable(t, r, c)

	takeControl(t, r, a, "c")
	takeControl(t, r, b, "c")

	_, err := r.ReleaseControl(a)
	require.NoError(t, err)
	require.Equal(t, remotecontrol.ModeTarget, lastDevices(r)[deviceID(t, "c")].Mode())

	r.Disconnect(b)
	require.Equal(t, remotecontrol.ModeLocal, lastDevices(r)[deviceID(t, "c")].Mode())
}

func TestRegistry_TargetGoingOfflineReleasesControllers(t *testing.T) {
	r := remotecontrol.NewRegistry()

	a := connect(t, r, "a", "Device A")
	setControllable(t, r, a)
	b := connect(t, r, "b", "Device B")
	setControllable(t, r, b)
	c := connect(t, r, "c", "Device C")
	setControllable(t, r, c)

	takeControl(t, r, a, "c")
	takeControl(t, r, b, "c")

	deliveries := r.Disconnect(c)

	require.Equal(t, remotecontrol.ReleaseReasonOffline, releasedReason(t, deliveries, a))
	require.Equal(t, remotecontrol.ReleaseReasonOffline, releasedReason(t, deliveries, b))

	devices := lastDevices(r)
	require.Equal(t, remotecontrol.ModeLocal, devices[deviceID(t, "a")].Mode())
	require.Equal(t, remotecontrol.ModeLocal, devices[deviceID(t, "b")].Mode())
	require.NotContains(t, devices, deviceID(t, "c"))
}

func TestRegistry_ClosingNonControllableTabOfTargetKeepsControllers(t *testing.T) {
	r := remotecontrol.NewRegistry()

	a := connect(t, r, "a", "Device A")
	setControllable(t, r, a)
	b := connect(t, r, "b", "Device B")
	setControllable(t, r, b)
	b2 := connect(t, r, "b", "Device B")

	takeControl(t, r, a, "b")

	deliveries := r.Disconnect(b2)
	for _, d := range deliveries {
		require.IsType(t, remotecontrol.DevicesEvent{}, d.Event())
	}

	require.Equal(t, remotecontrol.ModeTarget, lastDevices(r)[deviceID(t, "b")].Mode())
}

func TestRegistry_DisconnectControllers(t *testing.T) {
	r := remotecontrol.NewRegistry()

	a := connect(t, r, "a", "Device A")
	setControllable(t, r, a)
	b := connect(t, r, "b", "Device B")
	setControllable(t, r, b)
	c := connect(t, r, "c", "Device C")
	setControllable(t, r, c)

	takeControl(t, r, a, "c")
	takeControl(t, r, b, "c")

	deliveries, err := r.DisconnectControllers(c)
	require.NoError(t, err)

	require.Equal(t, remotecontrol.ReleaseReasonDisconnected, releasedReason(t, deliveries, a))
	require.Equal(t, remotecontrol.ReleaseReasonDisconnected, releasedReason(t, deliveries, b))

	require.Equal(t, remotecontrol.ModeLocal, lastDevices(r)[deviceID(t, "c")].Mode())
}

func TestRegistry_DisconnectControllersWithoutControllersDoesNothing(t *testing.T) {
	r := remotecontrol.NewRegistry()

	a := connect(t, r, "a", "Device A")
	setControllable(t, r, a)

	deliveries, err := r.DisconnectControllers(a)
	require.NoError(t, err)
	require.Empty(t, deliveries)
}

func TestRegistry_CommandIsSentToControllableConnectionOfTarget(t *testing.T) {
	r := remotecontrol.NewRegistry()

	a := connect(t, r, "a", "Device A")
	setControllable(t, r, a)
	b := connect(t, r, "b", "Device B")
	connect(t, r, "b", "Device B")
	setControllable(t, r, b)

	takeControl(t, r, a, "b")

	deliveries, err := r.Command(a, payload(t, "next"))
	require.NoError(t, err)
	require.Len(t, deliveries, 1)
	require.Equal(t, b, deliveries[0].Connection())
	command, ok := deliveries[0].Event().(remotecontrol.CommandEvent)
	require.True(t, ok)
	require.Equal(t, payload(t, "next"), command.Payload())
}

func TestRegistry_CommandRequiresController(t *testing.T) {
	r := remotecontrol.NewRegistry()

	a := connect(t, r, "a", "Device A")
	setControllable(t, r, a)

	_, err := r.Command(a, payload(t, "next"))
	require.ErrorIs(t, err, remotecontrol.ErrNotController)
}

func TestRegistry_StateIsSentToControllers(t *testing.T) {
	r := remotecontrol.NewRegistry()

	a := connect(t, r, "a", "Device A")
	setControllable(t, r, a)
	b := connect(t, r, "b", "Device B")
	setControllable(t, r, b)
	c := connect(t, r, "c", "Device C")
	setControllable(t, r, c)

	takeControl(t, r, a, "c")

	deliveries, err := r.State(c, payload(t, "state"))
	require.NoError(t, err)
	require.Len(t, deliveries, 1)
	require.Equal(t, a, deliveries[0].Connection())
	state, ok := deliveries[0].Event().(remotecontrol.StateEvent)
	require.True(t, ok)
	require.Equal(t, payload(t, "state"), state.Payload())
}

func TestRegistry_StateRequiresControllableConnection(t *testing.T) {
	r := remotecontrol.NewRegistry()

	a := connect(t, r, "a", "Device A")
	setControllable(t, r, a)
	a2 := connect(t, r, "a", "Device A")

	_, err := r.State(a2, payload(t, "state"))
	require.ErrorIs(t, err, remotecontrol.ErrNotControllable)
}

func TestRegistry_Rename(t *testing.T) {
	r := remotecontrol.NewRegistry()

	a := connect(t, r, "a", "Device A")
	setControllable(t, r, a)

	_, err := r.Rename(a, deviceName(t, "Renamed"))
	require.NoError(t, err)

	require.Equal(t, deviceName(t, "Renamed"), lastDevices(r)[deviceID(t, "a")].Name())
}

func TestRegistry_EmptyAfterLastDisconnect(t *testing.T) {
	r := remotecontrol.NewRegistry()

	a := connect(t, r, "a", "Device A")
	require.False(t, r.Empty())

	r.Disconnect(a)
	require.True(t, r.Empty())
}

func TestRegistry_UnknownConnection(t *testing.T) {
	r := remotecontrol.NewRegistry()
	unknown := remotecontrol.NewConnectionID()

	_, err := r.SetControllable(unknown)
	require.ErrorIs(t, err, remotecontrol.ErrConnectionNotFound)

	_, err = r.TakeControl(unknown, deviceID(t, "a"))
	require.ErrorIs(t, err, remotecontrol.ErrConnectionNotFound)

	require.Empty(t, r.Disconnect(unknown))
}

func connect(t *testing.T, r *remotecontrol.Registry, id, name string) remotecontrol.ConnectionID {
	connectionID := remotecontrol.NewConnectionID()
	_, err := r.Connect(connectionID, deviceID(t, id), deviceName(t, name))
	require.NoError(t, err)
	return connectionID
}

func setControllable(t *testing.T, r *remotecontrol.Registry, id remotecontrol.ConnectionID) {
	_, err := r.SetControllable(id)
	require.NoError(t, err)
}

func takeControl(t *testing.T, r *remotecontrol.Registry, id remotecontrol.ConnectionID, target string) {
	_, err := r.TakeControl(id, deviceID(t, target))
	require.NoError(t, err)
}

func lastDevices(r *remotecontrol.Registry) map[remotecontrol.DeviceID]remotecontrol.Device {
	devices := make(map[remotecontrol.DeviceID]remotecontrol.Device)
	for _, v := range r.Devices() {
		devices[v.ID()] = v
	}
	return devices
}

func listedDevices(r *remotecontrol.Registry) []remotecontrol.DeviceID {
	var ids []remotecontrol.DeviceID
	for _, v := range r.Devices() {
		ids = append(ids, v.ID())
	}
	return ids
}

func releasedReason(t *testing.T, deliveries []remotecontrol.Delivery, connection remotecontrol.ConnectionID) remotecontrol.ReleaseReason {
	var reasons []remotecontrol.ReleaseReason
	for _, d := range deliveries {
		if event, ok := d.Event().(remotecontrol.ReleasedEvent); ok && d.Connection() == connection {
			reasons = append(reasons, event.Reason())
		}
	}
	require.Len(t, reasons, 1)
	return reasons[0]
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
