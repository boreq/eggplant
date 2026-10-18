# Remote control: implementation plan

This plan implements `spec.md` in this directory.

## Assumptions

The spec does not decide these items yet. The plan uses the defaults below. Change them in `spec.md` first if you want something else.

- **Switching targets.** A controller that selects another device becomes a controller of that device directly.
- **Marking controllers.** The devices panel on a target marks its controllers.
- **Logout.** A logged-out device counts as offline.
- **Tabs in the panel.** The panel lists only the controllable tab of each browser.

## Naming

"Remote" already means a paired Eggplant instance (`domain/remote`, `application/remote`, `/api/remote`). The new backend packages use `remotecontrol` (`domain/remotecontrol`, `application/remotecontrol`, `adapters/remotecontrol`). Types and messages about one browser use `device`.

## Transport

One WebSocket per tab. WebSocket is better than SSE here for two reasons:

- Browsers allow only 6 HTTP/1.1 connections to one host, and an SSE stream holds one of them all the time. Many Eggplant installs use plain HTTP, so HTTP/2 is not available. With several tabs open, SSE streams block API requests and HLS fragments. WebSocket connections do not count toward this limit.
- The server knows which socket sent each message, so requests do not need to carry a connection ID.

## Model

- **Device**: one browser. The frontend generates a `DeviceID` once and keeps it in `localStorage`, together with the device name.
- **Connection**: one WebSocket, which is one tab. The backend generates a `ConnectionID` when the socket opens.
- A connection is controllable or not controllable. Each device has at most one controllable connection.
- A connection that controls another device stores the `DeviceID` of its target.
- The control rules apply at device level:
  - A device is a target if any connection controls it.
  - A device is a controller if any of its connections controls another device.
  - A device is never both.

## Backend

### Dependency

Add `github.com/coder/websocket`. The Go standard library has no WebSocket server.

### `domain/remotecontrol`

Pure code, no I/O.

- `DeviceID`: validated random ID from the client.
- `ConnectionID`: random ID from the backend.
- `DeviceName`: trimmed, 1 to 64 characters.
- `Registry`: the state for one user, with all control rules. Its methods return the events to send, so the rules can be unit-tested without sockets.
  - `Connect(connectionID, deviceID, name)`
  - `SetControllable(connectionID)`
  - `Disconnect(connectionID)`
  - `Rename(connectionID, name)`
  - `TakeControl(connectionID, targetDeviceID)`
  - `ReleaseControl(connectionID)`
  - `DisconnectControllers(connectionID)`
  - `Command(connectionID, command)`: routes the command to the controllable connection of the current target of the sender.
  - `State(connectionID, state)`: routes the state to the controllers of the device of the sender.
- Errors: `ErrConnectionNotFound`, `ErrConnectionExists`, `ErrDeviceNotFound`, `ErrCannotControlSelf`, `ErrDeviceIsTarget`, `ErrTargetIsController`, `ErrNotController`, `ErrNotControllable`.
- When a controllable connection disconnects, `Disconnect` releases all its controllers with the reason `offline`. `DisconnectControllers` does the same with the reason `disconnected`.

### `adapters/remotecontrol`

- `Hub`: keeps one `Registry` for each `Username` behind one mutex. The mutex makes "the first selection wins" true for two devices that select each other at the same time.
- Each connection gets one buffered outgoing channel. If a channel is full, the hub drops that connection, so a slow client cannot block the other clients.
- The hub is in memory only. It does not use Bolt or the Watermill pub/sub in `adapters/pubsub`, because those are persistent.

### `application/remotecontrol`

- New package with a `Control` struct of handlers, added to `application.Application`.
- No logging decorators from `genhandlers`. `application/auth` and `application/remote` do not use them either, and a target publishes its state every 5 seconds, which would flood the log.
- Handlers: `Connect`, `Disconnect`, `SetControllable`, `Rename`, `TakeControl`, `ReleaseControl`, `DisconnectControllers`, `SendCommand`, `PublishState`.
- `Connect` requires `accessctx.UserAccessContext`. Anonymous and paired-instance access contexts get `accessctx.ErrPermissionDenied`. The other handlers take the `ConnectionID` that `Connect` returned, so they do not check access again.
- The command and state payloads are opaque JSON for the backend. The backend passes them on and does not read them.

### `entrypoints/http`

New file `control.go` with one route: `GET /api/devices/ws?device=<id>&name=<name>`. The query parameters register the connection before the upgrade, so no `hello` message is needed.

- **Auth.** Run the existing `AuthProvider` on the upgrade request. The browser sends the `auth-token` cookie with the handshake. Reject anything that is not a user access context before the upgrade.
- **Origin check.** Keep the default origin check of `websocket.Accept`, which requires `Origin` to match `Host`. Cookies go with the handshake, so without this check any website can open a socket as the logged-in user. This attack is called cross-site WebSocket hijacking.
- **Read limit.** Call `SetReadLimit(2 << 20)`. The default limit is 32 KB, and a `replace` command with a full track list is larger.
- **Writer.** One goroutine reads the outgoing channel of the connection and writes to the socket. One loop reads from the socket and calls the handlers.
- **Keepalive.** Send a WebSocket ping every 15 seconds. A ping without a pong ends the connection.
- **Token check.** With each ping, check the access token again and close the socket if the token is no longer valid. This covers logout, password change, and user removal without new coupling between `application/auth` and `application/remotecontrol`.
- **Close.** Call `Disconnect` when the socket closes for any reason.
- `server.go`: route `/api/devices/ws` around `gziphandler`. `gziphandler` supports `Hijack`, so the upgrade works through it, but a bypass keeps the socket free of response wrappers.
- `http.Server` has no `WriteTimeout` today. Keep it that way, or long-lived sockets will break.

### Protocol

All messages are JSON objects with a `type` field. A client message with an `id` field gets one `result` message with the same `id`.

Client to server:

| Type | Fields | Purpose |
| --- | --- | --- |
| `controllable` | | The tab got the browser lock and is now the controllable tab of its device. |
| `rename` | `id`, `name` | Rename the device. |
| `take_control` | `id`, `device` | Take control of a device. |
| `release_control` | `id` | Return to local playback. |
| `disconnect_controllers` | `id` | The target disconnects its controllers. |
| `command` | `id`, `command` | A controller sends a command to its target. |
| `state` | `state` | The target publishes its playback state. |

Server to client:

| Type | Fields | Purpose |
| --- | --- | --- |
| `result` | `id`, `error` | The answer to a client message. `error` is empty on success. |
| `devices` | `devices` | The full device list, sent to all connections of the user after each change. Each entry has an ID, a name, a mode (`local`, `controller`, or `target`), and the IDs of its controllers. |
| `command` | `command` | A command for the target. |
| `state` | `state` | The playback state of the target, sent to its controllers. |
| `released` | `reason` | Control ended. The reason is `offline` or `disconnected`. |

Error values in `result`: `invalid_message`, `device_not_found`, `cannot_control_self`, `device_is_target`, `target_is_controller`, `not_controller`, `not_controllable`, `internal_error`.

### OpenAPI

- Add `GET /api/devices/ws` to `adapters/openapi/openapi.yaml` with a description of the upgrade and a link to the protocol section of this plan. OpenAPI cannot describe WebSocket messages.
- The generated client does not support WebSocket. Component tests use the `coder/websocket` client.

### Wire

- New `internal/wire/inject_control.go` with the hub and the handlers.
- Regenerate `wire_gen.go`.

### README

Add a note next to the `nginx` paragraph. A reverse proxy must pass the WebSocket upgrade for `/api/devices/ws`. For `nginx`, that means `proxy_http_version 1.1`, `proxy_set_header Upgrade $http_upgrade`, and `proxy_set_header Connection "upgrade"`.

## Frontend

### `vite.config.ts`

- Add `ws: true` to the `/api` proxy.
- `changeOrigin: true` changes `Host` but not `Origin`, so the backend origin check rejects the dev socket. Rewrite `Origin` to the proxy target in a `proxyReqWs` handler. The backend stays strict.

### `services/DeviceService.ts`

One instance, created in `App.ts`.

- **IDs.** Generate the `DeviceID` with `crypto.getRandomValues`. `crypto.randomUUID` needs a secure context, and many Eggplant installs use plain HTTP on a LAN.
- **Default name.** Build it from `navigator.userAgent`, for example `Firefox on Linux`.
- **Controllable tab.** Use `navigator.locks.request('eggplant-device', ...)` and hold the lock until the tab closes. The browser grants the lock in request order, so the first opened tab gets it, and the next oldest tab gets it when that tab closes. Send `controllable` when the lock arrives.
  - Web Locks also need a secure context. On plain HTTP, fall back to a `BroadcastChannel` election in which the tab with the oldest start time wins.
- **Socket.** Open the socket when `state.user` is set. Close it when `state.user` becomes `null`. Send `controllable` after each open if the tab holds the lock.
- **Reconnect.** Reconnect with backoff from 1 second up to 30 seconds. A closed socket ends control on the server, so a controller returns to local playback and shows a notification when its socket closes.
- **Requests.** Keep a map of pending `id` values to promises. Resolve or reject each promise when its `result` arrives.
- Expose the device list, the current mode, and the current target to the UI.

### `store.ts`

- New state fields: `devices`, `controlledDevice`, `deviceMode`.
- New mutations: `SetDevices`, `SetControlledDevice`, `SetRemoteState`. `SetRemoteState` replaces `entries`, `playingIndex`, `paused`, `volume`, `muted`, and `shuffle` with the state of the target.
- **Player mutations on a controller.** The player mutations (`Replace`, `Append`, `Remove`, `Play`, `Pause`, `Previous`, `Next`, `SetVolume`, `Mute`, `Unmute`, `Shuffle`, `NoShuffle`) do nothing while `controlledDevice` is set. Wrap them with one helper; do not edit each mutation.
- **Mutation forwarding.** `DeviceService` subscribes to the store mutations. On a controller, it sends each player mutation as a command. The target answers with its new state, which the controller applies through `SetRemoteState`.
  - The controller does not update its state before the answer. So `Next` with shuffle picks its random track on the target only, and the two devices never disagree.
  - `Remove` carries a `TrackWithAlbum` object, and the store finds it with `indexOf`. Object identity does not survive JSON. The plugin sends the index instead. The index is still correct because the mutation did nothing on the controller.
- **Persistence.** Add a `filter` to `vuex-persist` that skips persistence while `controlledDevice` is set. Without it, the volume of the target overwrites the saved local volume.
- **Take control.** Save the local `volume`, `muted`, and `shuffle`. Then pause and set `controlledDevice`.
- **Release.** Clear `controlledDevice`. Then replace the queue with an empty list, pause, and restore the saved `volume`, `muted`, and `shuffle`.

### `components/Player.ts`

- **On a controller:**
  - Tear down the stream when `controlledDevice` becomes set.
  - Ignore `storeNowPlaying`, `storePaused`, `storeVolume`, and `seekEvent` while `controlledDevice` is set.
- **On a target:**
  - Apply each incoming command with `store.commit`. A `remove` command maps its index back to the entry. A `seek` command emits `seekEvent` on `$root`.
  - Publish the state after each player mutation, with a debounce of 100 ms.
  - Publish the state every 5 seconds during playback, and after each seek.
  - The state contains `position` and the time of measurement, so that the controller can move its progress bar between updates.

### Seek and progress on a controller

- `Controls.ts` and `MediaSession.ts` emit `seekEvent`. `DeviceService` listens to it and sends a `seek` command while `controlledDevice` is set.
- `App.ts` gives `Controls` and `MediaSession` a `playbackData` value. On a controller, this value comes from `DeviceService`, which calculates the position from the last state plus the time since that state.

### UI

- `views/Browse.vue`: a new topbar button next to the queue button, visible only when `state.user` is set.
  - The icon changes with `deviceMode`. Pick three Font Awesome 5 icons, for example `fa-desktop` for local, `fa-gamepad` for controller, and `fa-broadcast-tower` for target.
- `components/Devices.vue`: a panel that the devices button opens in the content area of `Browse`, in the same way as the queue.
  - An editable name field for the current device.
  - The list of devices. Disable the entries that the rules block, and mark the controllers.
  - A click on the current device releases control on a controller, or disconnects the controllers on a target.
- Use `Notifications.pushError` for a rejected request, for `released` with the reason `offline`, and for `released` with the reason `disconnected`.

## Tests

- `domain/remotecontrol`: unit tests for each rule in `Registry`:
  - A controller cannot become a target.
  - A target cannot become a controller.
  - A device cannot control itself.
  - Several controllers can control one target.
  - Two devices that select each other: the first selection wins.
  - Release on disconnect and on `DisconnectControllers`.
  - A non-controllable connection cannot become a target.
- `adapters/remotecontrol`: a hub test for the event fan-out and for the drop of a slow connection.
- `internal/componenttest`, with the `coder/websocket` client:
  - Two logged-in clients. Client A takes control of client B, sends a command, gets the state from B, and gets `released` when B disconnects.
  - A handshake with a foreign `Origin` fails.
  - An anonymous handshake fails.
- Frontend: the project has no test runner. Run `yarn type-check` and `yarn lint`, and test by hand with two browsers and with two tabs in one browser.

## Order

- `domain/remotecontrol` and its tests.
- `adapters/remotecontrol` and `application/remotecontrol`.
- WebSocket endpoint, protocol, gzip bypass, OpenAPI, Wire, and README.
- Component tests.
- `vite.config.ts`, `DeviceService`, and the store changes.
- `Player.ts` changes for controllers and targets.
- Devices button and panel.
- Manual test with two browsers and two tabs.
