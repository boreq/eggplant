import { Component, Prop, Vue, Watch } from 'vue-property-decorator';
import { TrackWithAlbum } from '@/dto/TrackWithAlbum';
import { PlaybackData } from '@/dto/PlaybackData';
import { Mutation } from '@/store';
import { ApiService } from '@/services/ApiService';
import { seekEvent } from '@/components/Player';


@Component
export default class MediaSession extends Vue {

    @Prop()
    playbackData: PlaybackData;

    private readonly apiService = new ApiService(this);

    get nowPlaying(): TrackWithAlbum {
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

            const album = this.nowPlaying.album;
            const metadata = {
                title: this.nowPlaying.track.title,
                // album title is set as artist because the album field is unlikely to be displayed in mobile notifications
                artist: album ? album.title : null,
                album: album ? album.title : null,
                artwork: [
                    // hack to clear the artwork if a song doesn't have one
                    {
                        src: null,
                        sizes: null,
                    },
                ],
            };

            if (album && album.thumbnail) {
                const url = this.apiService.thumbnailUrl(album.thumbnail, album.remoteLibraryId);
                metadata.artwork = [
                    {
                        src: url,
                        sizes: '200x200',
                    },
                ];
            }

            mediaSession.metadata = new MediaMetadata(metadata);

            if (mediaSession.setPositionState) {
                try {
                    mediaSession.setPositionState({
                        duration: this.nowPlaying.track.duration,
                        position: 0,
                        playbackRate: 1,
                    });
                } catch {
                    // Some browsers throw when values are out of range.
                }
            }

            if (!this.$store.state.paused) {
                mediaSession.playbackState = 'playing';
            }
        }
    }

    @Watch('paused')
    onPausedChanged(): void {
        const mediaSession = this.mediaSession;
        if (mediaSession) {
            mediaSession.playbackState = this.paused ? 'paused' : 'playing';
        }
    }

    @Watch('playbackData')
    onPlaybackDataChanged(): void {
        const mediaSession = this.mediaSession;
        if (!mediaSession || !mediaSession.setPositionState) {
            return;
        }
        if (!this.playbackData) {
            return;
        }
        const duration = this.playbackData.duration;
        let position = this.playbackData.currentTime;
        if (!isFinite(position) || position < 0) {
            position = 0;
        }
        if (position > duration) {
            position = duration;
        }
        try {
            mediaSession.setPositionState({
                duration,
                position,
                playbackRate: 1,
            });
        } catch {
            // Some browsers throw when values are out of range.
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

            mediaSession.setActionHandler('seekto', (details: { seekTime: number }) => {
                this.onSeekTo(details.seekTime);
            });
        }
    }

    private onSeekTo(seconds: number): void {
        if (!this.playbackData || !this.playbackData.duration) {
            return;
        }
        this.$root.$emit(seekEvent, seconds / this.playbackData.duration);
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

