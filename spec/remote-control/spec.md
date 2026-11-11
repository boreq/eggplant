# Remote control

A logged-in user can control playback on any of their devices from any other device, like Spotify Connect.

## Terms

- **Device**: one browser in which the user is logged in.
- **Target**: a device that one or more controllers control.
- **Controller**: a device that sends commands to a target.

## User interface

- The topbar gets a devices button next to the queue button.
- The icon of the button shows the current mode of the device: local playback, controller of another device, or target of another device.
- A click on the button opens a devices panel. The panel lists all online devices of the current user.
- The panel shows the name of the current device. The user can edit this name.
- A click on another device makes the current device a controller of that device.
- On a controller, a click on the current device returns the current device to local playback.
- On a target, a click on the current device disconnects all its controllers.
- Anonymous users do not see the devices button.

## Device name

- The default device name comes from the user agent, for example `Firefox on Linux`.
- The user can rename the device.
- The device keeps its name across page reloads and new sessions in the same browser.

## Tabs

- If several tabs are open in one browser, only the first opened tab is controllable.
- When the controllable tab closes, the next oldest tab becomes controllable.
- The other tabs play audio locally and can become controllers. Other devices cannot control them.

## Behaviour

- Each device plays independently by default. Several devices can play audio at the same time.
- When a device becomes a controller, it pauses its own audio.
- When a controller returns to local playback, its queue is empty and playback is paused.
- The controller shows the state of the target: queue, current track, position, paused, volume, and shuffle.
- All player controls on the controller act on the target. They do not change local playback.
- More than one controller can control the same target. The last command wins.
- A device is never a controller and a target at the same time.
- A controller cannot become a target. The devices panel shows controllers, but the user cannot select them.
- A target cannot become a controller. While a device is a target, the user cannot select other devices in its devices panel.
- If two devices select each other at the same time, the first selection wins. The second selection fails and shows a notification.
- The target keeps its own controls.
- A device stops being a target when its last controller leaves. A controller leaves when it returns to local playback or goes offline.
- If the target goes offline or disconnects its controllers, all its controllers return to local playback and show a notification.
- A user can only see and control their own devices.
