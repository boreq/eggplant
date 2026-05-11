import { Component, Vue, Watch } from 'vue-property-decorator';
import { Entry } from '@/dto/Entry';
import { Mutation } from '@/store';
import { ApiService } from '@/services/ApiService';


@Component
export default class MediaSession extends Vue {

    private readonly apiService = new ApiService(this);

    get nowPlaying(): Entry {
        return this.$store.getters.nowPlaying;
    }

    get paused(): boolean {
        return !this.$store.getters.nowPlaying || this.$store.state.paused;
    }

    @Watch('nowPlaying')
    onNowPlayingChanged(): void {
        const mediaSession = this.mediaSession;
        if (mediaSession) {
            if (!this.nowPlaying) {
                mediaSession.metadata = null;
                return;
            }

            const metadata = {
                title: this.nowPlaying.track.title,
                album: this.nowPlaying.album.title,
                artwork: [
                    // hack to clear the artwork if a song doesn't have one
                    {
                        src: null,
                        sizes: null,
                    },
                ],
            };

            if (this.nowPlaying.album.thumbnail) {
                const url = this.apiService.thumbnailUrl(this.nowPlaying.album.thumbnail);
                metadata.artwork = [
                    {
                        src: url,
                        sizes: '200x200',
                    },
                ];
            }

            // @ts-ignore
            mediaSession.metadata = new MediaMetadata(metadata);
        }
    }

    @Watch('paused')
    onPausedChanged(): void {
        const mediaSession = this.mediaSession;
        if (mediaSession) {
            mediaSession.playbackState = this.paused ? 'paused' : 'playing';
        }
    }

    created(): void {
        const mediaSession = this.mediaSession;

        if (mediaSession) {
            // For some reason the type definitions are not supported?

            mediaSession.setActionHandler('play',
                () => this.onPlay(),
            );

            mediaSession.setActionHandler('pause',
                () => this.onPause(),
            );

            mediaSession.setActionHandler('previoustrack',
                () => this.onPrevious(),
            );

            mediaSession.setActionHandler('nexttrack',
                () => this.onNext(),
            );
        }
    }

    onPlay(): void {
        this.$store.commit(Mutation.Play);
    }

    onPause(): void {
        this.$store.commit(Mutation.Pause);
    }

    onPrevious(): void {
        this.$store.commit(Mutation.Previous);
    }

    onNext(): void {
        this.$store.commit(Mutation.Next);
    }

    private get mediaSession(): any {
        if ('mediaSession' in window.navigator) {
            return (window.navigator as any).mediaSession;
        }

        return null;
    }
}

