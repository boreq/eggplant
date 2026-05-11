import { Component, Vue, Prop, Watch } from 'vue-property-decorator';
import { Entry } from '@/dto/Entry';
import { BasicAlbum } from '@/dto/BasicAlbum';
import { SearchResult } from '@/dto/Search';
import { ApiService } from '@/services/ApiService';

import MainHeader from '@/components/MainHeader.vue';
import SubHeader from '@/components/SubHeader.vue';
import FormInput from '@/components/forms/FormInput.vue';
import AppButton from '@/components/forms/AppButton.vue';
import ActionBarButton from '@/components/ActionBarButton.vue';
import ActionBar from '@/components/ActionBar.vue';
import Tracks from '@/components/Tracks.vue';
import Albums from '@/components/Albums.vue';
import Spinner from '@/components/Spinner.vue';


@Component({
    components: {
        MainHeader,
        FormInput,
        AppButton,
        ActionBar,
        ActionBarButton,
        Tracks,
        Albums,
        SubHeader,
        Spinner,
    },
})
export default class Search extends Vue {

    @Prop()
    query: string;

    result: SearchResult = null;

    private timeoutId: number = null;
    private readonly apiService = new ApiService(this);
    private readonly searchDelay = 250;

    @Watch('query', { immediate: true })
    onQueryChanged(): void {
        this.scheduleTimeout();
    }

    get tracks(): Entry[] {
        if (!this.result || !this.result.tracks) {
            return [];
        }

        return this.result.tracks.map(
            v => {
                return {
                    album: v.album,
                    track: v.track,
                };
            },
        );
    }

    get albums(): BasicAlbum[] {
        if (!this.result || !this.result.albums) {
            return [];
        }

        return this.result.albums;
    }

    selectAlbum(event: BasicAlbum): void {
        this.$emit('select-album', event);
    }

    private load(): void {
        this.result = null;

        if (!this.query || this.query.length === 0) {
            return;
        }

        const query = this.query;

        this.apiService.search(this.query)
            .then(
                response => {
                    if (this.query === query) {
                        this.result = response.data;
                    }
                },
            );
    }

    private scheduleTimeout(): void {
        this.clearTimeout();
        this.timeoutId = window.setTimeout(this.load, this.searchDelay);
    }

    private clearTimeout(): void {
        if (this.timeoutId) {
            window.clearTimeout(this.timeoutId);
            this.timeoutId = null;
        }
    }

}
