import { Component, Vue } from 'vue-property-decorator';
import { ApiService } from '@/services/ApiService';
import { RemoteInstance } from '@/dto/RemoteInstance';
import Notifications from '@/components/Notifications';
import AppButton from '@/components/forms/AppButton.vue';
import FormInput from '@/components/forms/FormInput.vue';
import Spinner from '@/components/Spinner.vue';

const refreshIntervalMs = 3000;


@Component({
    components: {
        AppButton,
        FormInput,
        Spinner,
    },
})
export default class Remotes extends Vue {

    remotes: RemoteInstance[] = null;
    newRemoteUrl = '';
    adding = false;

    localPairingTokens: Record<string, string> = {};
    peerTokens: Record<string, string> = {};

    private readonly apiService = new ApiService(this);
    private refreshTimer: number = null;

    mounted(): void {
        this.load();
        this.refreshTimer = window.setInterval(() => this.load(true), refreshIntervalMs);
    }

    beforeDestroy(): void {
        if (this.refreshTimer !== null) {
            window.clearInterval(this.refreshTimer);
        }
    }

    addRemote(): void {
        if (!this.newRemoteUrl || this.adding) {
            return;
        }
        this.adding = true;
        this.apiService.addRemote(this.newRemoteUrl)
            .then(
                response => {
                    this.$set(this.localPairingTokens, response.data.id, response.data.local_pairing_token);
                    this.newRemoteUrl = '';
                    this.load();
                },
                error => {
                    Notifications.pushError(this, 'Could not connect the instance.', error);
                },
            ).finally(() => this.adding = false);
    }

    submitPeerToken(remote: RemoteInstance): void {
        const token = this.peerTokens[remote.id];
        if (!token) {
            return;
        }
        this.apiService.setRemotePairingToken(remote.id, token)
            .then(
                () => {
                    Notifications.pushSuccess(this, 'The pairing token has been submitted.');
                    this.$delete(this.peerTokens, remote.id);
                    this.load();
                },
                error => {
                    Notifications.pushError(this, 'Could not submit the pairing token.', error);
                },
            );
    }

    onPeerTokenInput(remote: RemoteInstance, value: string): void {
        this.$set(this.peerTokens, remote.id, value);
    }

    localPairingToken(remote: RemoteInstance): string {
        return this.localPairingTokens[remote.id];
    }

    statusText(remote: RemoteInstance): string {
        switch (remote.status) {
        case 'HEALTHY':
            return 'Connected.';
        case 'DEAD':
            return 'Paired, but the instance is currently unreachable.';
        case 'PAIRING':
            return remote.remote_pairing_token_set
                ? 'Finishing pairing, waiting for the other instance…'
                : 'Waiting for you to enter your friend\'s token.';
        default:
            return remote.status;
        }
    }

    statusClass(remote: RemoteInstance): string {
        switch (remote.status) {
        case 'HEALTHY':
            return 'good';
        case 'DEAD':
            return 'bad';
        default:
            return 'pending';
        }
    }

    copy(token: string): void {
        if (!navigator.clipboard) {
            Notifications.pushError(this, `
                The Clipboard API is not available.
                Please note that the Clipboard API is only available in secure
                contexts (websites secured with TLS).
            `);
            return;
        }
        navigator.clipboard.writeText(token)
            .then(
                () => {
                    Notifications.pushSuccess(this, 'The pairing token has been copied.');
                },
                () => {
                    Notifications.pushError(this, 'Copying to clipboard failed.');
                },
            );
    }

    private load(silent = false): void {
        this.apiService.listRemotes()
            .then(
                response => {
                    this.remotes = response.data;
                },
                error => {
                    if (!silent) {
                        Notifications.pushError(this, 'Could not list the connected instances.', error);
                    }
                },
            );
    }

}
