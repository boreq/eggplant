package http

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/boreq/eggplant/application/accessctx"
	"github.com/boreq/eggplant/application/auth"
	appremotecontrol "github.com/boreq/eggplant/application/remotecontrol"
	authdomain "github.com/boreq/eggplant/domain/auth"
	"github.com/boreq/eggplant/domain/remotecontrol"
	"github.com/boreq/errors"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/julienschmidt/httprouter"
)

const (
	deviceReadLimit    = 2 << 20
	devicePingInterval = 15 * time.Second
	devicePingTimeout  = 10 * time.Second
)

var errInvalidDeviceMessage = errors.New("invalid message")

type deviceClientMessage struct {
	Type    string          `json:"type"`
	ID      string          `json:"id"`
	Name    string          `json:"name"`
	Device  string          `json:"device"`
	Command json.RawMessage `json:"command"`
	State   json.RawMessage `json:"state"`
}

type deviceResultMessage struct {
	Type  string `json:"type"`
	ID    string `json:"id"`
	Error string `json:"error,omitempty"`
}

type deviceDevicesMessage struct {
	Type    string      `json:"type"`
	Devices []deviceDTO `json:"devices"`
}

type deviceDTO struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Mode        string   `json:"mode"`
	Controllers []string `json:"controllers"`
}

type deviceCommandMessage struct {
	Type    string          `json:"type"`
	Command json.RawMessage `json:"command"`
}

type deviceStateMessage struct {
	Type  string          `json:"type"`
	State json.RawMessage `json:"state"`
}

type deviceReleasedMessage struct {
	Type   string `json:"type"`
	Reason string `json:"reason"`
}

func (h *Handler) deviceWebsocket(accessCtx accessctx.AccessContext, w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	userCtx, ok := accessCtx.(accessctx.UserAccessContext)
	if !ok {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	deviceID, err := remotecontrol.NewDeviceID(r.URL.Query().Get("device"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	name, err := remotecontrol.NewDeviceName(r.URL.Query().Get("name"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ws, err := websocket.Accept(w, r, nil)
	if err != nil {
		h.log.Debug("could not accept the websocket", "err", err)
		return
	}
	defer ws.CloseNow()

	conn, err := h.app.Control.Connect.Execute(accessCtx, appremotecontrol.Connect{DeviceID: deviceID, Name: name})
	if err != nil {
		if errors.Is(err, accessctx.ErrPermissionDenied) {
			_ = ws.Close(websocket.StatusPolicyViolation, "permission denied")
			return
		}
		h.log.Error("could not connect the device", "err", err)
		_ = ws.Close(websocket.StatusInternalError, "could not connect the device")
		return
	}
	defer h.app.Control.Disconnect.Execute(appremotecontrol.Disconnect{Connection: conn.ID})

	ws.SetReadLimit(deviceReadLimit)

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	var wg sync.WaitGroup
	wg.Go(func() {
		defer cancel()
		h.deviceWriteLoop(ctx, ws, conn.Events)
	})
	wg.Go(func() {
		defer cancel()
		h.deviceKeepAliveLoop(ctx, ws, userCtx.Token())
	})

	h.deviceReadLoop(ctx, ws, conn.ID)
	cancel()
	wg.Wait()
}

func (h *Handler) deviceReadLoop(ctx context.Context, ws *websocket.Conn, id remotecontrol.ConnectionID) {
	for {
		typ, data, err := ws.Read(ctx)
		if err != nil {
			return
		}

		if typ != websocket.MessageText {
			continue
		}

		var msg deviceClientMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			h.log.Debug("could not decode a device message", "err", err)
			continue
		}

		err = h.handleDeviceMessage(id, msg)
		if err != nil && deviceErrorCode(err) == "internal_error" {
			h.log.Error("could not handle a device message", "type", msg.Type, "err", err)
		}

		if msg.ID == "" {
			continue
		}

		result := deviceResultMessage{Type: "result", ID: msg.ID}
		if err != nil {
			result.Error = deviceErrorCode(err)
		}

		if err := wsjson.Write(ctx, ws, result); err != nil {
			return
		}
	}
}

func (h *Handler) deviceWriteLoop(ctx context.Context, ws *websocket.Conn, events <-chan remotecontrol.Event) {
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-events:
			if !ok {
				_ = ws.Close(websocket.StatusTryAgainLater, "connection dropped")
				return
			}

			msg, err := toDeviceServerMessage(event)
			if err != nil {
				h.log.Error("could not convert a device event", "err", err)
				continue
			}

			if err := wsjson.Write(ctx, ws, msg); err != nil {
				return
			}
		}
	}
}

func (h *Handler) deviceKeepAliveLoop(ctx context.Context, ws *websocket.Conn, token authdomain.AccessToken) {
	ticker := time.NewTicker(devicePingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, devicePingTimeout)
			err := ws.Ping(pingCtx)
			cancel()
			if err != nil {
				return
			}

			if _, err := h.app.Auth.CheckAccessToken.Execute(auth.CheckAccessToken{Token: token}); err != nil {
				if errors.Is(err, auth.ErrUnauthorized) {
					_ = ws.Close(websocket.StatusPolicyViolation, "unauthorized")
					return
				}
				h.log.Error("could not check the access token", "err", err)
			}
		}
	}
}

func (h *Handler) handleDeviceMessage(id remotecontrol.ConnectionID, msg deviceClientMessage) error {
	switch msg.Type {
	case "controllable":
		return h.app.Control.SetControllable.Execute(appremotecontrol.SetControllable{Connection: id})
	case "rename":
		name, err := remotecontrol.NewDeviceName(msg.Name)
		if err != nil {
			return errors.Wrap(errInvalidDeviceMessage, err.Error())
		}
		return h.app.Control.Rename.Execute(appremotecontrol.Rename{Connection: id, Name: name})
	case "take_control":
		target, err := remotecontrol.NewDeviceID(msg.Device)
		if err != nil {
			return errors.Wrap(errInvalidDeviceMessage, err.Error())
		}
		return h.app.Control.TakeControl.Execute(appremotecontrol.TakeControl{Connection: id, Target: target})
	case "release_control":
		return h.app.Control.ReleaseControl.Execute(appremotecontrol.ReleaseControl{Connection: id})
	case "disconnect_controllers":
		return h.app.Control.DisconnectControllers.Execute(appremotecontrol.DisconnectControllers{Connection: id})
	case "command":
		payload, err := remotecontrol.NewPayload(msg.Command)
		if err != nil {
			return errors.Wrap(errInvalidDeviceMessage, err.Error())
		}
		return h.app.Control.SendCommand.Execute(appremotecontrol.SendCommand{Connection: id, Payload: payload})
	case "state":
		payload, err := remotecontrol.NewPayload(msg.State)
		if err != nil {
			return errors.Wrap(errInvalidDeviceMessage, err.Error())
		}
		return h.app.Control.PublishState.Execute(appremotecontrol.PublishState{Connection: id, Payload: payload})
	default:
		return errors.Wrap(errInvalidDeviceMessage, "unknown message type")
	}
}

func deviceErrorCode(err error) string {
	switch {
	case errors.Is(err, errInvalidDeviceMessage):
		return "invalid_message"
	case errors.Is(err, remotecontrol.ErrDeviceNotFound):
		return "device_not_found"
	case errors.Is(err, remotecontrol.ErrCannotControlSelf):
		return "cannot_control_self"
	case errors.Is(err, remotecontrol.ErrDeviceIsTarget):
		return "device_is_target"
	case errors.Is(err, remotecontrol.ErrTargetIsController):
		return "target_is_controller"
	case errors.Is(err, remotecontrol.ErrNotController):
		return "not_controller"
	case errors.Is(err, remotecontrol.ErrNotControllable):
		return "not_controllable"
	default:
		return "internal_error"
	}
}

func toDeviceServerMessage(event remotecontrol.Event) (any, error) {
	switch e := event.(type) {
	case remotecontrol.DevicesEvent:
		devices := make([]deviceDTO, 0, len(e.Devices()))
		for _, d := range e.Devices() {
			controllers := make([]string, 0, len(d.Controllers()))
			for _, c := range d.Controllers() {
				controllers = append(controllers, c.String())
			}
			devices = append(devices, deviceDTO{
				ID:          d.ID().String(),
				Name:        d.Name().String(),
				Mode:        d.Mode().String(),
				Controllers: controllers,
			})
		}
		return deviceDevicesMessage{Type: "devices", Devices: devices}, nil
	case remotecontrol.CommandEvent:
		return deviceCommandMessage{Type: "command", Command: e.Payload().Bytes()}, nil
	case remotecontrol.StateEvent:
		return deviceStateMessage{Type: "state", State: e.Payload().Bytes()}, nil
	case remotecontrol.ReleasedEvent:
		return deviceReleasedMessage{Type: "released", Reason: e.Reason().String()}, nil
	default:
		return nil, errors.New("unknown event")
	}
}
