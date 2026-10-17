export enum DeviceMode {
    Local = 'local',
    Controller = 'controller',
    Target = 'target',
}

export class Device {
    id: string;
    name: string;
    mode: DeviceMode;
    controllers: string[];
}
