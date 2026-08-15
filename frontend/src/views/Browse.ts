import { Component, Vue, Ref, Watch } from 'vue-property-decorator';
import { ApiService } from '@/services/ApiService';
import { DurationLoader } from '@/services/DurationLoader';
import { HttpStatus } from '@/services/HttpStatus';
import { NavigationService } from '@/services/NavigationService';
import { Album, PartialAlbum } from '@/dto/Album';
import { Library } from '@/dto/Library';
import { User } from '@/dto/User';
import { TrackWithAlbum } from '@/dto/TrackWithAlbum';
import { Track } from '@/dto/Track';
import { Mutation, ReplaceCommand, AppendCommand } from '@/store';
import { Location } from 'vue-router';
import Notifications from '@/components/Notifications';

import SubHeader from '@/components/SubHeader.vue';
import MainHeader from '@/components/MainHeader.vue';
import Albums from '@/components/Albums.vue';
import RemoteLibraries from '@/components/RemoteLibraries.vue';
import Tracks from '@/components/Tracks.vue';
import Thumbnail from '@/components/Thumbnail.vue';
import NowPlaying from '@/components/NowPlaying.vue';
import LoginButton from '@/components/LoginButton.vue';
import SearchInput from '@/components/forms/SearchInput.vue';
import Queue from '@/components/Queue.vue';
import Search from '@/components/Search.vue';
import Dropdown from '@/components/Dropdown.vue';
import DropdownElement from '@/components/DropdownElement.vue';
import DropdownDivider from '@/components/DropdownDivider.vue';
import Spinner from '@/components/Spinner.vue';

enum View {
    Browse = 'browse',
    Search = 'search',
    Queue = 'queue',
}

enum BrowseState {
    Loading = 'loading',
    Ready = 'ready',
    Empty = 'empty',
    PermissionDeniedOrDoesNotExist = 'permission-denied',
    LibraryNotReady = 'library-not-ready',
}

@Component({
    components: {
        Albums,
        RemoteLibraries,
        SubHeader,
        MainHeader,
        Tracks,
        Thumbnail,
        NowPlaying,
        SearchInput,
        LoginButton,
        Queue,
        Search,
        Dropdown,
        DropdownElement,
        DropdownDivider,
        Spinner,
    },
})
export default class Browse extends Vue {

    album: Album = null;
    remoteLibraries: Library[] = null;
    state: BrowseState = BrowseState.Loading;

    searchQuery: string = null;

    view: View = View.Browse;

    highlightTrackId: string | null = null;

    showAllTracks: boolean = false;
    showAllAlbums: boolean = false;

    private readonly initialTrackLimit = 5;
    private readonly initialAlbumLimit = 5;

    @Ref('dropdown')
    readonly dropdown: Dropdown;

    @Ref('content')
    readonly contentDiv: HTMLDivElement;

    private timeoutId: number;

    private readonly apiService = new ApiService(this);
    private readonly navigationService = new NavigationService();
    private readonly durationLoader = new DurationLoader(this);

    @Watch('album')
    onAlbumLoaded(): void {
        if (this.highlightTrackId) {
            this.doReveal();
        }
    }

    @Watch('$route')
    onRouteChanged(): void {
        this.album = null;
        this.state = BrowseState.Loading;
        this.showAllTracks = false;
        this.showAllAlbums = false;
        this.load();
        this.scrollToTop();
        this.switchView(View.Browse);
    }

    @Watch('user')
    onUserChanged(): void {
        this.loadRemoteLibraries();
    }

    @Watch('searchQuery')
    onSearchQueryChanged(): void {
        if (this.searchQuery) {
            if (this.view !== View.Search) {
                this.switchView(View.Search);
            }
        } else {
            if (this.view === View.Search) {
                this.switchView(View.Browse);
            }
        }
    }

    created(): void {
        this.load();
        this.loadRemoteLibraries();
        this.$root.$on('revealNowPlaying', this.revealNowPlaying);
        this.$root.$on('scrollToTop', this.scrollToTop);
    }

    destroyed(): void {
        this.$root.$off('revealNowPlaying', this.revealNowPlaying);
        this.$root.$off('scrollToTop', this.scrollToTop);
        this.clearTimeout();
        this.durationLoader.cancel();
    }

    parentUrl(album: PartialAlbum): Location {
        return this.navigationService.getBrowse(album.id, album.remoteInstanceId);
    }

    selectAlbum(album: PartialAlbum): void {
        this.switchView(View.Browse);
        const alreadyOnPage = this.$route.params.albumId === album.id
            && this.$route.params.instanceId === (album.remoteInstanceId || undefined);
        if (alreadyOnPage) {
            return;
        }
        const location = this.navigationService.getBrowse(album.id, album.remoteInstanceId);
        this.$router.push(location);
    }

    selectRemoteLibrary(library: Library): void {
        this.switchView(View.Browse);
        const location = this.navigationService.getBrowse(undefined, library.id);
        this.$router.push(location);
    }

    toggleQueue(): void {
        if (this.view === View.Queue) {
            this.switchView(View.Browse);
        } else {
            this.switchView(View.Queue);
        }
    }

    onPlayAlbumButtonClicked(): void {
        if (this.anyAlbumSongIsCurrentlyNowPlaying()) {
            if (this.paused()) {
                this.$store.commit(Mutation.Play);
            } else {
                this.$store.commit(Mutation.Pause);
            }
        } else {
            const entries: TrackWithAlbum[] = this.entries
                .map(v => {
                    return {
                        album: v.album,
                        track: v.track,
                    };
                });
            const command: ReplaceCommand = {
                entries: entries,
                playingIndex: 0,
            };
            this.$store.commit(Mutation.Replace, command);
        }
    }

    addAlbumToQueue(): void {
        const entries: TrackWithAlbum[] = this.entries
            .map(v => {
                return {
                    album: v.album,
                    track: v.track,
                };
            });
        const command: AppendCommand = {
            entries: entries,
        };
        this.$store.commit(Mutation.Append, command);
        this.dropdown.close();
        Notifications.pushSuccess(this, 'Album added to queue.');
    }

    onSearchNavigation(): void {
        this.scrollToTop();
        this.searchQuery = null;
    }

    private revealNowPlaying(trackId: string): void {
        this.highlightTrackId = trackId;
        if (this.album) {
            this.doReveal();
        }
    }

    private doReveal(): void {
        this.showAllTracks = true;
        this.$nextTick(() => {
            const playingEl = this.contentDiv.querySelector('.track.playing > :first-child') as HTMLElement | null;
            if (playingEl) {
                playingEl.scrollIntoView({ behavior: 'smooth', block: 'center' });
            }
            window.setTimeout(() => {
                this.highlightTrackId = null;
            }, 2500);
        });
    }

    get showSearch(): boolean {
        return this.view === View.Search;
    }

    get showAlbum(): boolean {
        if (this.album) {
            if (!this.getIdFromRoute()) {
                return !!this.getInstanceFromRoute()
                    || !!this.album.thumbnail
                    || (this.album.tracks && this.album.tracks.length > 0);
            }
            return true;
        }
        return false;
    }

    get title(): string {
        if (this.album && this.album.title) {
            return this.album.title;
        }
        if (this.remoteInstanceId) {
            return this.remoteInstanceTitle;
        }
        return 'Eggplant';
    }

    get remoteInstanceId(): string {
        return this.getInstanceFromRoute();
    }

    get remoteInstanceTitle(): string {
        const library = (this.remoteLibraries || [])
            .find(v => v.id === this.remoteInstanceId);
        return library ? library.name : 'Remote library';
    }

    get remoteInstanceUrl(): Location {
        return this.navigationService.getBrowse(undefined, this.remoteInstanceId);
    }

    get showRemoteLibraries(): boolean {
        return !this.getIdFromRoute()
            && !this.getInstanceFromRoute()
            && !!this.remoteLibraries
            && this.remoteLibraries.length > 0;
    }

    get user(): User {
        return this.$store.state.user;
    }

    get showQueue(): boolean {
        return this.view === View.Queue;
    }

    get noContent(): boolean {
        if (!this.album) {
            return false;
        }

        if (this.showRemoteLibraries) {
            return false;
        }

        if (this.album.tracks && this.album.tracks.length > 0) {
            return false;
        }

        if (this.album.albums && this.album.albums.length > 0) {
            return false;
        }

        return true;
    }

    get numberOfTracks(): number {
        if (this.album && this.album.tracks) {
            return this.album.tracks.length;
        }
        return 0;
    }

    get entries(): TrackWithAlbum[] {
        if (this.album && this.album.tracks) {
            return this.album.tracks
                .map((v: Track): TrackWithAlbum => {
                    return {
                        track: v,
                        album: this.toPartialAlbum(this.album),
                    };
                });
        }
        return [];
    }

    get basicAlbum(): PartialAlbum {
        return this.toPartialAlbum(this.album);
    }

    get albums(): PartialAlbum[] {
        if (!this.album || !this.album.albums) {
            return null;
        }
        return this.album.albums;
    }

    get hasAlbums(): boolean {
        return !!(this.albums && this.albums.length > 0);
    }

    get hasTracks(): boolean {
        return !!(this.album && this.album.tracks && this.album.tracks.length > 0);
    }

    get displayedEntries(): TrackWithAlbum[] {
        if (this.hasAlbums && !this.showAllTracks) {
            return this.entries.slice(0, this.initialTrackLimit);
        }
        return this.entries;
    }

    get canToggleTracks(): boolean {
        return this.hasAlbums && this.entries.length > this.initialTrackLimit;
    }

    get displayedAlbums(): PartialAlbum[] {
        if (!this.albums) {
            return null;
        }
        if (this.hasTracks && !this.showAllAlbums) {
            return this.albums.slice(0, this.initialAlbumLimit);
        }
        return this.albums;
    }

    get canToggleAlbums(): boolean {
        return this.hasTracks
            && !!this.albums
            && this.albums.length > this.initialAlbumLimit;
    }

    get totalDurationMinutes(): number {
        if (this.album && this.album.tracks) {
            return Math.ceil(
                this.album.tracks.reduce(
                    (acc, track) => track.duration ? acc + track.duration : acc,
                    0,
                ) / 60,
            );
        }
        return 0;
    }

    get durationsLoading(): boolean {
        if (this.album && this.album.tracks) {
            return this.album.tracks.some(track => track.duration === undefined);
        }
        return false;
    }

    get showPlayAlbumButtonAsPause(): boolean {
        if (!this.anyAlbumSongIsCurrentlyNowPlaying()) {
            return false;
        }

        return !this.paused();
    }

    get nowPlaying(): TrackWithAlbum {
        return this.$store.getters.nowPlaying;
    }

    private paused(): boolean {
        return this.$store.getters.nowPlayingPaused;
    }

    private anyAlbumSongIsCurrentlyNowPlaying(): boolean {
        if (!this.nowPlaying) {
            return false;
        }

        for (const track of this.album.tracks) {
            if (track.id === this.nowPlaying.track.id) {
                return true;
            }
        }

        return false;
    }

    private load(): void {
        this.clearTimeout();
        this.apiService.browse(this.getIdFromRoute(), this.getInstanceFromRoute())
            .then(
                response => {
                    this.album = response.data;
                    this.state = this.noContent ? BrowseState.Empty : BrowseState.Ready;

                    if (this.album && this.album.tracks) {
                        this.durationLoader.load(this.album.tracks);
                    }

                    if (this.state === BrowseState.Empty) {
                        this.scheduleTimeout();
                    }
                },
                error => {
                    const status = error.response && error.response.status;
                    if (status === HttpStatus.ServiceUnavailable) {
                        this.state = BrowseState.LibraryNotReady;
                        this.scheduleTimeout();
                        return;
                    }

                    Notifications.pushError(this, 'Could not list the tracks and albums.', error);

                    if (status === HttpStatus.Forbidden || status === HttpStatus.NotFound) {
                        this.state = BrowseState.PermissionDeniedOrDoesNotExist;
                        return;
                    }

                    this.scheduleTimeout();
                });
    }

    private loadRemoteLibraries(): void {
        if (!this.user) {
            this.remoteLibraries = null;
            return;
        }

        this.apiService.listLibraries()
            .then(
                response => {
                    this.remoteLibraries = response.data;
                    if (this.state === BrowseState.Empty && !this.noContent) {
                        this.state = BrowseState.Ready;
                    }
                },
                error => {
                    this.remoteLibraries = null;
                    Notifications.pushError(this, 'Could not list the remote libraries.', error);
                });
    }

    private scheduleTimeout(): void {
        this.clearTimeout();
        this.timeoutId = window.setTimeout(this.load, 5000);
    }

    private clearTimeout(): void {
        if (this.timeoutId) {
            window.clearTimeout(this.timeoutId);
            this.timeoutId = null;
        }
    }

    private getIdFromRoute(): string | undefined {
        return this.$route.params.albumId;
    }

    private getInstanceFromRoute(): string | undefined {
        const instance = this.$route.params.instanceId;
        return typeof instance === 'string' ? instance : undefined;
    }

    scrollToTop(smooth: boolean = false): void {
        this.contentDiv.scrollTo({ top: 0, behavior: smooth ? 'smooth' : 'auto' });
    }

    private toPartialAlbum(album: Album): PartialAlbum {
        return {
            title: album.title,
            id: album.id,
            thumbnail: album.thumbnail,
            remoteInstanceId: album.remoteInstanceId,
        };
    }

    private switchView(view: View): void {
        if (view !== View.Search) {
            this.searchQuery = null;
        }

        this.view = view;
    }

}
