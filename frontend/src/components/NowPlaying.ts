import { Component, Vue } from 'vue-property-decorator';
import Thumbnail from '@/components/Thumbnail.vue';
import { TrackWithAlbum } from '@/dto/TrackWithAlbum';
import { NavigationService } from '@/services/NavigationService';


@Component({
    components: {
        Thumbnail,
    },
})
export default class NowPlaying extends Vue {

    private readonly navigationService = new NavigationService();

    get nowPlaying(): TrackWithAlbum {
        return this.$store.getters.nowPlaying;
    }

    get trackTitle(): string {
        if (this.nowPlaying && this.nowPlaying.track) {
            return this.nowPlaying.track.title;
        }
        return null;
    }

    get albumTitle(): string {
        if (this.nowPlaying && this.nowPlaying.album) {
            return this.nowPlaying.album.title;
        }
        return null;
    }

    goToNowPlayingSong(): void {
        this.navigateToNowPlaying(true);
    }

    goToNowPlayingAlbum(): void {
        this.navigateToNowPlaying(false);
    }

    private navigateToNowPlaying(reveal: boolean): void {
        const { id: albumId, remoteLibraryId } = this.nowPlaying.album;
        const trackId = this.nowPlaying.track.id;
        const location = this.navigationService.getBrowse(albumId, remoteLibraryId);
        const currentAlbumId = this.$route.params.albumId;
        const currentLibraryId = this.$route.params.libraryId;

        const alreadyOnPage = currentAlbumId === albumId && currentLibraryId === (remoteLibraryId || undefined);
        if (alreadyOnPage) {
            if (reveal) {
                this.$root.$emit('revealNowPlaying', trackId);
            } else {
                this.$root.$emit('scrollToTop');
            }
        } else {
            this.$router.push(location).then(() => {
                if (reveal) {
                    this.$root.$emit('revealNowPlaying', trackId);
                } else {
                    this.$root.$emit('scrollToTop');
                }
            });
        }
    }

}
