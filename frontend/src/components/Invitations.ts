import { Component, Vue } from 'vue-property-decorator';
import { ApiService } from '@/services/ApiService';
import Notifications from '@/components/Notifications';
import AppButton from '@/components/forms/AppButton.vue';


@Component({
    components: {
        AppButton,
    },
})
export default class Invitation extends Vue {

    working = false;
    private token: string = null;
    private readonly apiService = new ApiService(this);

    createInvitation(): void {
        this.working = true;
        this.apiService.createInvitation()
            .then(
                response => {
                    this.token = response.data.token;
                },
                error => {
                    Notifications.pushError(this, 'Generating an invitation failed.', error);
                },
            ).finally(() => this.working = false);
    }

    copy(): void {
        if (!navigator.clipboard) {
            Notifications.pushError(this, `
                The Clipboard API is not available.
                Please note that the Clipboard API is only available in secure
                contexts (websites secured with TLS).
            `);
            return;
        }

        navigator.clipboard.writeText(this.invitationUrl)
            .then(
                () => {
                    Notifications.pushSuccess(this, 'The invitation link has been copied.');
                },
                () => {
                    Notifications.pushError(this, 'Copying to clipboard failed.');
                },
            );
    }

    get invitationUrl(): string {
        if (!this.token) {
            return null;
        }
        const l = window.location;
        return `${l.protocol}//${l.host}/register/${this.token}`;
    }

}
