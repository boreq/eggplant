import { Component, Vue } from 'vue-property-decorator';
import MainHeader from '@/components/MainHeader.vue';
import ActionBarButton from '@/components/ActionBarButton.vue';
import ActionBar from '@/components/ActionBar.vue';
import Notifications from '@/components/Notifications';
import { ApiService } from '@/services/ApiService';
import { Stats as StatsDto } from '@/dto/Stats';
import Spinner from '@/components/Spinner.vue';
import StoreStats from '@/components/StoreStats.vue';
import SubHeader from '@/components/SubHeader.vue';


@Component({
    components: {
        MainHeader,
        SubHeader,
        ActionBar,
        ActionBarButton,
        Spinner,
        StoreStats,
    },
})
export default class Stats extends Vue {

    stats: StatsDto = null;

    private timeoutID: number;

    private readonly apiService = new ApiService(this);

    mounted(): void {
        this.load();
    }

    destroyed(): void {
        window.clearTimeout(this.timeoutID);
    }

    goToSettings(): void {
        this.$router.push({name: 'settings'});
    }

    private load(): void {
        this.apiService.stats()
            .then(
                response => {
                    this.stats = response.data;
                },
                error => {
                    Notifications.pushError(this, 'Could not retrieve the statistics.', error);
                },
            )
            .finally(() => {
                this.timeoutID = window.setTimeout(this.load, 10 * 1000);
            });
    }

}
