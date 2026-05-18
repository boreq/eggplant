import { Component, Vue } from 'vue-property-decorator';
import { ApiService } from '@/services/ApiService';
import Notifications from '@/components/Notifications';
import Spinner from '@/components/Spinner.vue';


@Component({
    components: {
        Spinner,
    },
})
export default class Version extends Vue {

    backendVersion: string | null = null;

    private readonly apiService = new ApiService(this);

    get frontendVersion(): string {
        return import.meta.env.VUE_APP_VERSION || 'unknown';
    }

    mounted(): void {
        this.apiService.getVersion()
            .then(
                response => {
                    this.backendVersion = response.data.version;
                },
                error => {
                    Notifications.pushError(this, 'Could not fetch the backend version.', error);
                },
            );
    }

}
