import { Component, Vue } from 'vue-property-decorator';
import { ApiService } from '@/services/ApiService';
import Notifications from '@/components/Notifications';
import Spinner from '@/components/Spinner.vue';
import { Version as VersionDto } from '@/dto/Version';


@Component({
    components: {
        Spinner,
    },
})
export default class Version extends Vue {

    version: VersionDto | null = null;

    private readonly apiService = new ApiService(this);

    get backendVersion(): string | null {
        return this.version ? this.version.backend : null;
    }

    get frontendVersion(): string | null {
        return this.version ? this.version.frontend : null;
    }

    mounted(): void {
        this.apiService.getVersion()
            .then(
                response => {
                    this.version = response.data;
                },
                error => {
                    Notifications.pushError(this, 'Could not fetch the backend version.', error);
                },
            );
    }

}
