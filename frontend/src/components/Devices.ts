import { Component, Vue } from 'vue-property-decorator';
import { Device, DeviceMode } from '@/dto/Device';
import { deviceErrorMessage, deviceModeIcon, deviceService } from '@/services/DeviceService';
import Notifications from '@/components/Notifications';

import SubHeader from '@/components/SubHeader.vue';
import FormInput from '@/components/forms/FormInput.vue';
import AppButton from '@/components/forms/AppButton.vue';


@Component({
    components: {
        SubHeader,
        FormInput,
        AppButton,
    },
})
export default class Devices extends Vue {

    name = deviceService.deviceName;
    renaming = false;

    get devices(): Device[] {
        return this.$store.state.devices;
    }

    get controlledDevice(): string {
        return this.$store.state.controlledDevice;
    }

    get mode(): DeviceMode {
        return deviceService.mode;
    }

    get currentName(): string {
        const device = deviceService.ownDevice;
        return device ? device.name : deviceService.deviceName;
    }

    get nameChanged(): boolean {
        const trimmed = this.name ? this.name.trim() : '';
        return trimmed !== '' && trimmed !== this.currentName;
    }

    rename(): void {
        if (!this.nameChanged || this.renaming) {
            return;
        }
        this.renaming = true;
        deviceService.rename(this.name)
            .then(() => {
                Notifications.pushSuccess(this, 'Device renamed.');
            })
            .catch(err => {
                Notifications.pushError(this, deviceErrorMessage(err));
            })
            .finally(() => {
                this.renaming = false;
            });
    }

    isOwn(device: Device): boolean {
        return device.id === deviceService.deviceId;
    }

    isCurrent(device: Device): boolean {
        if (this.controlledDevice) {
            return device.id === this.controlledDevice;
        }
        return this.isOwn(device);
    }

    canSelect(device: Device): boolean {
        if (this.isOwn(device)) {
            return !!this.controlledDevice || this.mode === DeviceMode.Target;
        }
        if (device.id === this.controlledDevice) {
            return false;
        }
        if (this.mode === DeviceMode.Target) {
            return false;
        }
        return device.mode !== DeviceMode.Controller;
    }

    select(device: Device): void {
        if (!this.canSelect(device)) {
            return;
        }

        let request: Promise<void>;
        if (!this.isOwn(device)) {
            request = deviceService.takeControl(device.id);
        } else if (this.controlledDevice) {
            request = deviceService.releaseControl();
        } else {
            request = deviceService.disconnectControllers();
        }

        request.catch(err => {
            Notifications.pushError(this, deviceErrorMessage(err));
        });
    }

    icon(device: Device): string {
        if (this.isOwn(device) && this.controlledDevice) {
            return deviceModeIcon(DeviceMode.Controller);
        }
        return deviceModeIcon(device.mode);
    }

    describe(device: Device): string {
        const parts: string[] = [];

        if (this.isOwn(device)) {
            parts.push('This device');
        }

        if (device.id === this.controlledDevice) {
            parts.push('You control this device');
        } else if (device.mode === DeviceMode.Target) {
            parts.push('Controlled by ' + this.controllerNames(device));
        } else if (device.mode === DeviceMode.Controller) {
            parts.push('Controls another device');
        } else {
            parts.push('Plays on its own');
        }

        if (this.isOwn(device) && this.controlledDevice) {
            parts.push('Click to play on this device again');
        } else if (this.isOwn(device) && device.mode === DeviceMode.Target) {
            parts.push('Click to disconnect the controllers');
        }

        return parts.join(' · ');
    }

    private controllerNames(device: Device): string {
        return device.controllers
            .map(id => {
                const controller = this.devices.find(d => d.id === id);
                return controller ? controller.name : 'another device';
            })
            .join(', ');
    }

}
