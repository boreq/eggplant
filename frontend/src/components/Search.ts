import { Component, Vue, Prop, Watch } from 'vue-property-decorator';
import { TrackWithAlbum } from '@/dto/TrackWithAlbum';
import { ApiService } from '@/services/ApiService';
import { DurationLoader } from '@/services/DurationLoader';

import MainHeader from '@/components/MainHeader.vue';
import SubHeader from '@/components/SubHeader.vue';
import FormInput from '@/components/forms/FormInput.vue';
import AppButton from '@/components/forms/AppButton.vue';
import ActionBarButton from '@/components/ActionBarButton.vue';
import ActionBar from '@/components/ActionBar.vue';
import Tracks from '@/components/Tracks.vue';
import Albums from '@/components/Albums.vue';
import Spinner from '@/components/Spinner.vue';
import { SearchResults } from "@/dto/Search";
import { PartialAlbum } from "@/dto/Album";


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

    result: SearchResults = null;

    showAllTracks: boolean = false;
    showAllAlbums: boolean = false;

    private readonly initialTrackLimit = 5;
    private readonly initialAlbumLimit = 5;

    private timeoutId: number = null;
    private readonly apiService = new ApiService(this);
    private readonly durationLoader = new DurationLoader(this);
    private readonly searchDelay = 50;

    destroyed(): void {
        this.clearTimeout();
        this.durationLoader.cancel();
    }

    @Watch('query', { immediate: true })
    onQueryChanged(): void {
        this.showAllTracks = false;
        this.showAllAlbums = false;
        this.scheduleTimeout();
    }

    get tracks(): TrackWithAlbum[] {
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

    get albums(): PartialAlbum[] {
        if (!this.result || !this.result.albums) {
            return [];
        }

        return this.result.albums.map(a => a.album);
    }

    get hasTracks(): boolean {
        return this.tracks.length > 0;
    }

    get hasAlbums(): boolean {
        return this.albums.length > 0;
    }

    get displayedTracks(): TrackWithAlbum[] {
        if (!this.showAllTracks) {
            return this.tracks.slice(0, this.initialTrackLimit);
        }
        return this.tracks;
    }

    get canToggleTracks(): boolean {
        return this.tracks.length > this.initialTrackLimit;
    }

    get displayedAlbums(): PartialAlbum[] {
        if (!this.showAllAlbums) {
            return this.albums.slice(0, this.initialAlbumLimit);
        }
        return this.albums;
    }

    get canToggleAlbums(): boolean {
        return this.albums.length > this.initialAlbumLimit;
    }

    selectAlbum(event: PartialAlbum): void {
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
                        if (this.result && this.result.tracks) {
                            this.durationLoader.load(this.result.tracks.map(entry => entry.track));
                        }
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
