import { Component, Vue } from 'vue-property-decorator';
import { ApiService } from '@/services/ApiService';
import { RemoteInstance, AddRemoteCommand } from '@/dto/RemoteInstance';
import Notifications from '@/components/Notifications';
import FormInput from '@/components/forms/FormInput.vue';
import AppButton from '@/components/forms/AppButton.vue';

@Component({
    components: {
        FormInput,
        AppButton,
    },
})
export default class LinkedInstances extends Vue {

    instances: RemoteInstance[] = [];
    loading = true;

    addName = '';
    addUrl = '';
    addUsername = '';
    addPassword = '';
    addInProgress = false;

    private readonly apiService = new ApiService(this);

    created(): void {
        this.load();
    }

    load(): void {
        this.loading = true;
        this.apiService.listRemotes()
            .then(
                response => {
                    this.instances = response.data;
                },
                error => {
                    Notifications.pushError(this, 'Could not load linked instances.', error);
                },
            )
            .finally(() => {
                this.loading = false;
            });
    }

    add(): void {
        if (!this.addName || !this.addUrl || !this.addUsername || !this.addPassword) {
            return;
        }
        this.addInProgress = true;
        const cmd: AddRemoteCommand = {
            name: this.addName,
            url: this.addUrl,
            username: this.addUsername,
            password: this.addPassword,
        };
        this.apiService.addRemote(cmd)
            .then(
                response => {
                    this.instances.push(response.data);
                    this.addName = '';
                    this.addUrl = '';
                    this.addUsername = '';
                    this.addPassword = '';
                    this.$emit('instances-changed', this.instances);
                },
                error => {
                    Notifications.pushError(this, 'Could not link instance.', error);
                },
            )
            .finally(() => {
                this.addInProgress = false;
            });
    }

    remove(instance: RemoteInstance): void {
        this.apiService.removeRemote(instance.id)
            .then(
                () => {
                    this.instances = this.instances.filter(i => i.id !== instance.id);
                    this.$emit('instances-changed', this.instances);
                },
                error => {
                    Notifications.pushError(this, 'Could not unlink instance.', error);
                },
            );
    }

}
