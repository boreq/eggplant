import Vue from 'vue';
import store, { Mutation, PlaybackState, RemoveCommand, playerMutations } from '@/store';
import { Device, DeviceMode } from '@/dto/Device';
import { PlaybackData } from '@/dto/PlaybackData';
import { TrackWithAlbum } from '@/dto/TrackWithAlbum';
import { seekEvent } from '@/components/Player';
import Notifications from '@/components/Notifications';

const deviceIdKey = 'eggplant_device_id';
const deviceNameKey = 'eggplant_device_name';
const lockName = 'eggplant_device';
const electionChannelName = 'eggplant_device_election';

const maxDeviceNameLength = 64;

const reconnectMinDelay = 1_000;
const reconnectMaxDelay = 30_000;
const statePublishDelay = 100;
const statePublishInterval = 5_000;
const seekStatePublishDelay = 1_000;
const electionHeartbeatInterval = 1_000;
const electionTimeout = 3_000;

type DeviceCommand =
    { type: 'mutation', mutation: string, payload?: any } |
    { type: 'seek', position: number };

class RemoteState {
    entries: TrackWithAlbum[];
    playingIndex: number;
    paused: boolean;
    volume: number;
    muted: boolean;
    shuffle: boolean;
    position: number;
    duration: number;
}

class LocalSettings {
    volume: number;
    muted: boolean;
    shuffle: boolean;
}

interface PendingRequest {
    resolve: () => void;
    reject: (error: DeviceError) => void;
}

export class DeviceError extends Error {
    constructor(readonly code: string) {
        super(code);
    }
}

export class DeviceService {

    readonly deviceId = loadOrCreate(deviceIdKey, randomId);

    private vue: Vue = null;
    private socket: WebSocket = null;
    private controllable = false;
    private reconnectAttempt = 0;
    private reconnectTimeoutID: number = null;
    private requestCounter = 0;
    private readonly pendingRequests = new Map<string, PendingRequest>();

    private localSettings: LocalSettings = null;
    private takingControl = false;
    private pendingState: RemoteState = null;
    private remotePlayback: { position: number, duration: number, paused: boolean, receivedAt: number } = null;
    private localPlaybackData: PlaybackData = null;

    private publishTimeoutID: number = null;
    private publishIntervalID: number = null;
    private knownControllers: string[] = [];

    get deviceName(): string {
        try {
            const name = window.localStorage.getItem(deviceNameKey);
            return isValidDeviceName(name) ? name.trim() : defaultDeviceName();
        } catch {
            return defaultDeviceName();
        }
    }

    start(vue: Vue): void {
        this.vue = vue;

        this.electControllableTab();

        store.watch(state => state.user, user => {
            if (user) {
                this.connect();
            } else {
                this.disconnect();
            }
        }, { immediate: true });

        store.subscribe(mutation => this.onMutation(mutation.type, mutation.payload));

        vue.$root.$on(seekEvent, (position: number) => {
            if (store.state.controlledDevice) {
                this.sendCommand({ type: 'seek', position });
            }
        });
    }

    get mode(): DeviceMode {
        if (store.state.controlledDevice) {
            return DeviceMode.Controller;
        }
        const device = this.ownDevice;
        return device ? device.mode : DeviceMode.Local;
    }

    get ownDevice(): Device {
        return store.state.devices.find(device => device.id === this.deviceId);
    }

    rename(name: string): Promise<void> {
        if (!isValidDeviceName(name)) {
            return Promise.reject(new DeviceError('invalid_name'));
        }
        const trimmed = name.trim();
        return this.request({ type: 'rename', name: trimmed }).then(() => {
            try {
                window.localStorage.setItem(deviceNameKey, trimmed);
            } catch {
                throw new DeviceError('storage_unavailable');
            }
        });
    }

    takeControl(id: string): Promise<void> {
        this.takingControl = true;
        this.pendingState = null;
        return this.request({ type: 'take_control', device: id }).then(() => {
            if (!store.state.controlledDevice) {
                this.localSettings = {
                    volume: store.state.volume,
                    muted: store.state.muted,
                    shuffle: store.state.shuffle,
                };
                store.commit(Mutation.Pause);
            }
            store.commit(Mutation.SetControlledDevice, id);
            store.commit(Mutation.SetPlaybackState, emptyPlaybackState(this.localSettings));
            this.remotePlayback = null;
            if (this.pendingState) {
                this.applyRemoteState(this.pendingState);
            }
        }).finally(() => {
            this.takingControl = false;
            this.pendingState = null;
        });
    }

    releaseControl(): Promise<void> {
        return this.request({ type: 'release_control' }).finally(() => this.returnToLocalPlayback());
    }

    disconnectControllers(): Promise<void> {
        return this.request({ type: 'disconnect_controllers' });
    }

    setLocalPlaybackData(playbackData: PlaybackData): void {
        this.localPlaybackData = playbackData;
    }

    remotePlaybackData(): PlaybackData {
        if (!this.remotePlayback) {
            return null;
        }
        const elapsed = this.remotePlayback.paused ? 0 : (Date.now() - this.remotePlayback.receivedAt) / 1000;
        let currentTime = this.remotePlayback.position + elapsed;
        if (this.remotePlayback.duration && currentTime > this.remotePlayback.duration) {
            currentTime = this.remotePlayback.duration;
        }
        return {
            currentTime: currentTime,
            duration: this.remotePlayback.duration,
        };
    }

    private connect(): void {
        if (this.socket) {
            return;
        }

        window.clearTimeout(this.reconnectTimeoutID);

        const url = new URL(import.meta.env.VUE_APP_API_PREFIX + 'devices/ws', window.location.href);
        url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:';
        url.searchParams.set('device', this.deviceId);
        url.searchParams.set('name', this.deviceName);

        const socket = new WebSocket(url.toString());
        this.socket = socket;

        socket.onopen = () => {
            if (this.controllable) {
                this.send({ type: 'controllable' });
            }
        };

        socket.onmessage = (event: MessageEvent) => {
            this.onMessage(JSON.parse(event.data));
        };

        socket.onclose = () => {
            if (this.socket !== socket) {
                return;
            }
            this.socket = null;
            this.onConnectionLost();
            if (store.state.user) {
                this.scheduleReconnect();
            }
        };
    }

    private disconnect(): void {
        window.clearTimeout(this.reconnectTimeoutID);
        if (!this.socket) {
            return;
        }
        const socket = this.socket;
        this.socket = null;
        socket.close();
        this.onConnectionLost();
    }

    private scheduleReconnect(): void {
        const delay = Math.min(reconnectMaxDelay, reconnectMinDelay * Math.pow(2, this.reconnectAttempt));
        this.reconnectAttempt++;
        this.reconnectTimeoutID = window.setTimeout(() => this.connect(), delay);
    }

    private onConnectionLost(): void {
        for (const pending of this.pendingRequests.values()) {
            pending.reject(new DeviceError('connection_closed'));
        }
        this.pendingRequests.clear();

        store.commit(Mutation.SetDevices, []);
        this.updateControllers([]);

        if (store.state.controlledDevice) {
            this.returnToLocalPlayback();
            Notifications.pushError(this.vue, 'Lost the connection to the controlled device.');
        }
    }

    private onMessage(msg: any): void {
        switch (msg.type) {
            case 'result':
                this.onResult(msg.id, msg.error);
                break;
            case 'devices':
                this.reconnectAttempt = 0;
                this.onDevices(msg.devices);
                break;
            case 'command':
                this.onCommand(msg.command);
                break;
            case 'state':
                this.onState(msg.state);
                break;
            case 'released':
                this.onReleased(msg.reason);
                break;
        }
    }

    private onResult(id: string, error: string): void {
        const pending = this.pendingRequests.get(id);
        if (!pending) {
            return;
        }
        this.pendingRequests.delete(id);
        if (error) {
            pending.reject(new DeviceError(error));
        } else {
            pending.resolve();
        }
    }

    private onDevices(devices: Device[]): void {
        store.commit(Mutation.SetDevices, devices);
        const device = this.ownDevice;
        this.updateControllers(this.controllable && device ? device.controllers : []);
    }

    private onCommand(command: DeviceCommand): void {
        if (!this.controllable) {
            return;
        }

        if (command.type === 'seek') {
            this.vue.$root.$emit(seekEvent, command.position);
            this.schedulePublishState(seekStatePublishDelay);
            return;
        }

        if (command.type !== 'mutation' || !playerMutations.includes(command.mutation)) {
            return;
        }

        if (command.mutation === Mutation.Remove) {
            const entry = store.state.entries[command.payload.index];
            if (entry) {
                const removeCommand: RemoveCommand = { entry };
                store.commit(Mutation.Remove, removeCommand);
            }
            return;
        }

        store.commit(command.mutation, command.payload);
    }

    private onState(state: RemoteState): void {
        if (this.takingControl) {
            this.pendingState = state;
            return;
        }
        if (store.state.controlledDevice) {
            this.applyRemoteState(state);
        }
    }

    private onReleased(reason: string): void {
        if (!store.state.controlledDevice) {
            return;
        }
        const target = store.state.devices.find(device => device.id === store.state.controlledDevice);
        const name = target ? `"${target.name}"` : 'The controlled device';
        this.returnToLocalPlayback();
        if (reason === 'disconnected') {
            Notifications.pushError(this.vue, `${name} disconnected you.`);
        } else {
            Notifications.pushError(this.vue, `${name} went offline.`);
        }
    }

    private onMutation(type: string, payload: any): void {
        if (!playerMutations.includes(type)) {
            return;
        }

        if (store.state.controlledDevice) {
            if (type === Mutation.Remove) {
                const index = store.state.entries.indexOf((payload as RemoveCommand).entry);
                if (index >= 0) {
                    this.sendCommand({ type: 'mutation', mutation: type, payload: { index } });
                }
                return;
            }
            this.sendCommand({ type: 'mutation', mutation: type, payload });
            return;
        }

        if (this.knownControllers.length > 0) {
            this.schedulePublishState(statePublishDelay);
        }
    }

    private applyRemoteState(state: RemoteState): void {
        const playbackState: PlaybackState = {
            entries: state.entries,
            playingIndex: state.playingIndex,
            paused: state.paused,
            volume: state.volume,
            muted: state.muted,
            shuffle: state.shuffle,
        };
        store.commit(Mutation.SetPlaybackState, playbackState);
        this.remotePlayback = {
            position: state.position || 0,
            duration: state.duration,
            paused: state.paused,
            receivedAt: Date.now(),
        };
    }

    private returnToLocalPlayback(): void {
        if (!store.state.controlledDevice) {
            return;
        }
        store.commit(Mutation.SetControlledDevice, null);
        store.commit(Mutation.SetPlaybackState, emptyPlaybackState(this.localSettings));
        this.localSettings = null;
        this.remotePlayback = null;
    }

    private updateControllers(controllers: string[]): void {
        const added = controllers.some(id => !this.knownControllers.includes(id));
        this.knownControllers = controllers;

        if (controllers.length === 0) {
            window.clearInterval(this.publishIntervalID);
            this.publishIntervalID = null;
            return;
        }

        if (this.publishIntervalID === null) {
            this.publishIntervalID = window.setInterval(() => {
                if (!store.state.paused) {
                    this.publishState();
                }
            }, statePublishInterval);
        }

        if (added) {
            this.publishState();
        }
    }

    private schedulePublishState(delay: number): void {
        window.clearTimeout(this.publishTimeoutID);
        this.publishTimeoutID = window.setTimeout(() => this.publishState(), delay);
    }

    private publishState(): void {
        if (!this.controllable || this.knownControllers.length === 0) {
            return;
        }
        const nowPlaying: TrackWithAlbum = store.getters.nowPlaying;
        const state: RemoteState = {
            entries: store.state.entries,
            playingIndex: store.state.playingIndex,
            paused: store.state.paused,
            volume: store.state.volume,
            muted: store.state.muted,
            shuffle: store.state.shuffle,
            position: nowPlaying && this.localPlaybackData ? this.localPlaybackData.currentTime : 0,
            duration: nowPlaying ? nowPlaying.track.duration : null,
        };
        this.send({ type: 'state', state });
    }

    private sendCommand(command: DeviceCommand): void {
        this.request({ type: 'command', command }).catch((err: DeviceError) => {
            Notifications.pushError(this.vue, `Could not control the device (${err.code}).`);
        });
    }

    private request(msg: object): Promise<void> {
        if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
            return Promise.reject(new DeviceError('connection_closed'));
        }
        const id = String(++this.requestCounter);
        return new Promise<void>((resolve, reject) => {
            this.pendingRequests.set(id, { resolve, reject });
            this.send({ ...msg, id });
        });
    }

    private send(msg: object): void {
        if (this.socket && this.socket.readyState === WebSocket.OPEN) {
            this.socket.send(JSON.stringify(msg));
        }
    }

    private onBecameControllable(): void {
        this.controllable = true;
        this.send({ type: 'controllable' });
    }

    private electControllableTab(): void {
        const locks = (navigator as any).locks;
        if (locks && locks.request) {
            locks.request(lockName, () => {
                this.onBecameControllable();
                return new Promise(() => {});
            });
            return;
        }

        if (typeof BroadcastChannel === 'undefined') {
            this.onBecameControllable();
            return;
        }

        this.electWithBroadcastChannel();
    }

    private electWithBroadcastChannel(): void {
        const channel = new BroadcastChannel(electionChannelName);
        const tabId = randomId();
        const startedAt = Date.now();
        const lastSeen = new Map<string, { startedAt: number, seenAt: number }>();

        const isOlder = (otherId: string, otherStartedAt: number) =>
            otherStartedAt < startedAt || (otherStartedAt === startedAt && otherId < tabId);

        channel.onmessage = (event: MessageEvent) => {
            const msg = event.data;
            if (msg.type === 'bye') {
                lastSeen.delete(msg.tabId);
            } else {
                lastSeen.set(msg.tabId, { startedAt: msg.startedAt, seenAt: Date.now() });
            }
        };

        window.addEventListener('beforeunload', () => {
            channel.postMessage({ type: 'bye', tabId });
        });

        const heartbeat = () => channel.postMessage({ type: 'alive', tabId, startedAt });
        heartbeat();

        const intervalID = window.setInterval(() => {
            heartbeat();
            const now = Date.now();
            for (const [otherId, other] of lastSeen) {
                if (now - other.seenAt > electionTimeout) {
                    lastSeen.delete(otherId);
                }
            }
            const olderTabExists = Array.from(lastSeen.entries())
                .some(([otherId, other]) => isOlder(otherId, other.startedAt));
            if (!olderTabExists && now - startedAt > electionTimeout) {
                window.clearInterval(intervalID);
                window.setInterval(heartbeat, electionHeartbeatInterval);
                this.onBecameControllable();
            }
        }, electionHeartbeatInterval);
    }
}

function isValidDeviceName(name: string): boolean {
    if (!name) {
        return false;
    }
    const trimmed = name.trim();
    return trimmed.length > 0 && Array.from(trimmed).length <= maxDeviceNameLength;
}

function emptyPlaybackState(settings: LocalSettings): PlaybackState {
    return {
        entries: [],
        playingIndex: null,
        paused: true,
        volume: settings ? settings.volume : store.state.volume,
        muted: settings ? settings.muted : store.state.muted,
        shuffle: settings ? settings.shuffle : store.state.shuffle,
    };
}

function loadOrCreate(key: string, create: () => string): string {
    try {
        const value = window.localStorage.getItem(key);
        if (value) {
            return value;
        }
        const created = create();
        window.localStorage.setItem(key, created);
        return created;
    } catch {
        return create();
    }
}

function randomId(): string {
    const bytes = new Uint8Array(16);
    crypto.getRandomValues(bytes);
    return Array.from(bytes, b => b.toString(16).padStart(2, '0')).join('');
}

function defaultDeviceName(): string {
    const ua = navigator.userAgent;

    let browser = 'Browser';
    if (/Edg\//.test(ua)) {
        browser = 'Edge';
    } else if (/OPR\//.test(ua)) {
        browser = 'Opera';
    } else if (/Firefox\//.test(ua)) {
        browser = 'Firefox';
    } else if (/Chrome\//.test(ua)) {
        browser = 'Chrome';
    } else if (/Safari\//.test(ua)) {
        browser = 'Safari';
    }

    let os = 'Unknown OS';
    if (/Android/.test(ua)) {
        os = 'Android';
    } else if (/iPhone|iPad|iPod/.test(ua)) {
        os = 'iOS';
    } else if (/Windows/.test(ua)) {
        os = 'Windows';
    } else if (/Mac OS X/.test(ua)) {
        os = 'macOS';
    } else if (/CrOS/.test(ua)) {
        os = 'ChromeOS';
    } else if (/Linux/.test(ua)) {
        os = 'Linux';
    }

    return `${browser} on ${os}`;
}

export const deviceService = new DeviceService();

export function deviceModeIcon(mode: DeviceMode): string {
    switch (mode) {
        case DeviceMode.Controller:
            return 'fa-gamepad';
        case DeviceMode.Target:
            return 'fa-broadcast-tower';
        default:
            return 'fa-desktop';
    }
}

export function deviceErrorMessage(err: unknown): string {
    const code = err instanceof DeviceError ? err.code : null;
    switch (code) {
        case 'device_not_found':
            return 'This device is no longer online.';
        case 'cannot_control_self':
            return 'A device cannot control itself.';
        case 'device_is_target':
            return 'This device is controlled by another device, so it cannot control other devices.';
        case 'target_is_controller':
            return 'That device controls another device, so it cannot be controlled.';
        case 'invalid_name':
            return `The device name must have 1 to ${maxDeviceNameLength} characters.`;
        case 'storage_unavailable':
            return 'The device was renamed, but the browser could not save the name.';
        case 'invalid_message':
            return 'The server rejected the request as invalid.';
        case 'connection_closed':
            return 'There is no connection to the server.';
        default:
            return 'The request failed.';
    }
}
