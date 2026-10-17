package componenttest_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/stretchr/testify/require"
)

func TestServiceDeviceRemoteControl(t *testing.T) {
	ts := newTestService(t)
	token := registerAdminAndLogin(t, ts)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	a := dialDevice(t, ctx, ts, token, "device-a", "Device A")
	b := dialDevice(t, ctx, ts, token, "device-b", "Device B")

	writeDeviceMessage(t, ctx, a, map[string]any{"type": "controllable"})
	writeDeviceMessage(t, ctx, b, map[string]any{"type": "controllable"})

	waitForDeviceMessage(t, ctx, a, func(msg map[string]any) bool {
		return msg["type"] == "devices" && len(msg["devices"].([]any)) == 2
	})

	writeDeviceMessage(t, ctx, b, map[string]any{"type": "take_control", "id": "1", "device": "device-b"})
	result := waitForDeviceMessage(t, ctx, b, isDeviceResult("1"))
	require.Equal(t, "cannot_control_self", result["error"])

	writeDeviceMessage(t, ctx, a, map[string]any{"type": "take_control", "id": "2", "device": "device-b"})
	result = waitForDeviceMessage(t, ctx, a, isDeviceResult("2"))
	require.Nil(t, result["error"])

	writeDeviceMessage(t, ctx, b, map[string]any{"type": "take_control", "id": "3", "device": "device-a"})
	result = waitForDeviceMessage(t, ctx, b, isDeviceResult("3"))
	require.Equal(t, "device_is_target", result["error"])

	writeDeviceMessage(t, ctx, a, map[string]any{"type": "command", "id": "4", "command": map[string]any{"type": "next"}})
	command := waitForDeviceMessage(t, ctx, b, isDeviceMessage("command"))
	require.Equal(t, map[string]any{"type": "next"}, command["command"])

	writeDeviceMessage(t, ctx, b, map[string]any{"type": "state", "state": map[string]any{"paused": true}})
	state := waitForDeviceMessage(t, ctx, a, isDeviceMessage("state"))
	require.Equal(t, map[string]any{"paused": true}, state["state"])

	require.NoError(t, b.Close(websocket.StatusNormalClosure, ""))

	released := waitForDeviceMessage(t, ctx, a, isDeviceMessage("released"))
	require.Equal(t, "offline", released["reason"])
}

func TestServiceDeviceHandshake(t *testing.T) {
	ts := newTestService(t)
	token := registerAdminAndLogin(t, ts)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("anonymous handshake fails", func(t *testing.T) {
		_, resp, err := websocket.Dial(ctx, deviceURL(ts, "device-a", "Device A"), &websocket.DialOptions{
			HTTPHeader: http.Header{"Origin": []string{ts.baseURL}},
		})
		require.Error(t, err)
		require.NotNil(t, resp)
		defer resp.Body.Close()
		require.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("handshake with a foreign origin fails", func(t *testing.T) {
		_, resp, err := websocket.Dial(ctx, deviceURL(ts, "device-a", "Device A"), &websocket.DialOptions{
			HTTPHeader: http.Header{
				"Origin": []string{"http://evil.example.com"},
				"Cookie": []string{"auth-token=" + token},
			},
		})
		require.Error(t, err)
		require.NotNil(t, resp)
		defer resp.Body.Close()
		require.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("handshake with a foreign origin does not change the device", func(t *testing.T) {
		a := dialDevice(t, ctx, ts, token, "device-origin", "Device A")
		writeDeviceMessage(t, ctx, a, map[string]any{"type": "controllable"})

		_, resp, err := websocket.Dial(ctx, deviceURL(ts, "device-origin", "Evil"), &websocket.DialOptions{
			HTTPHeader: http.Header{
				"Origin": []string{"http://evil.example.com"},
				"Cookie": []string{"auth-token=" + token},
			},
		})
		require.Error(t, err)
		require.NotNil(t, resp)
		defer resp.Body.Close()

		writeDeviceMessage(t, ctx, a, map[string]any{"type": "rename", "id": "1", "name": "Renamed"})

		waitForDeviceMessage(t, ctx, a, func(msg map[string]any) bool {
			if msg["type"] != "devices" {
				return false
			}
			for _, d := range msg["devices"].([]any) {
				name := d.(map[string]any)["name"]
				require.NotEqual(t, "Evil", name)
				if name == "Renamed" {
					return true
				}
			}
			return false
		})
	})

	t.Run("handshake without a device id fails", func(t *testing.T) {
		_, resp, err := websocket.Dial(ctx, deviceURL(ts, "", "Device A"), &websocket.DialOptions{
			HTTPHeader: http.Header{
				"Origin": []string{ts.baseURL},
				"Cookie": []string{"auth-token=" + token},
			},
		})
		require.Error(t, err)
		require.NotNil(t, resp)
		defer resp.Body.Close()
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func dialDevice(t *testing.T, ctx context.Context, ts *testService, token, id, name string) *websocket.Conn {
	conn, resp, err := websocket.Dial(ctx, deviceURL(ts, id, name), &websocket.DialOptions{
		HTTPHeader: http.Header{
			"Origin": []string{ts.baseURL},
			"Cookie": []string{"auth-token=" + token},
		},
	})
	require.NoError(t, err)
	if resp.Body != nil {
		_ = resp.Body.Close()
	}
	t.Cleanup(func() {
		_ = conn.CloseNow()
	})
	return conn
}

func deviceURL(ts *testService, id, name string) string {
	query := url.Values{}
	query.Set("device", id)
	query.Set("name", name)
	return "ws" + strings.TrimPrefix(ts.baseURL, "http") + "/api/devices/ws?" + query.Encode()
}

func writeDeviceMessage(t *testing.T, ctx context.Context, conn *websocket.Conn, msg map[string]any) {
	require.NoError(t, wsjson.Write(ctx, conn, msg))
}

func waitForDeviceMessage(t *testing.T, ctx context.Context, conn *websocket.Conn, match func(msg map[string]any) bool) map[string]any {
	for {
		_, data, err := conn.Read(ctx)
		require.NoError(t, err)

		var msg map[string]any
		require.NoError(t, json.Unmarshal(data, &msg))

		if match(msg) {
			return msg
		}
	}
}

func isDeviceMessage(typ string) func(msg map[string]any) bool {
	return func(msg map[string]any) bool {
		return msg["type"] == typ
	}
}

func isDeviceResult(id string) func(msg map[string]any) bool {
	return func(msg map[string]any) bool {
		return msg["type"] == "result" && msg["id"] == id
	}
}
