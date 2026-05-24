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
        this.goToNowPlayingAlbum();
    }

    goToNowPlayingAlbum(): void {
        const albumId = this.nowPlaying.album.id;
        if (this.$route.params.id === albumId) {
            return;
        }
        this.$router.push(this.navigationService.getBrowse(albumId));
    }

}
