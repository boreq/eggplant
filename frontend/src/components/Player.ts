import { Component, Vue, Watch } from 'vue-property-decorator';
import Hls from 'hls.js';
import { Mutation } from '@/store';
import { ApiService } from '@/services/ApiService';
import { PlaybackData } from '@/dto/PlaybackData';
import { Entry } from '@/dto/Entry';
import { Track } from '@/dto/Track';
import Notifications from '@/components/Notifications';

export const seekEvent = 'seek';

@Component({
    components: {},
})
export default class Player extends Vue {

    private intervalID: number;

    private readonly apiService = new ApiService(this);

    private currentNowPlaying: Entry = null;

    private hls: Hls = null;

    private streamWs: WebSocket = null;

    private streamStartOffset = 0;

    get storeNowPlaying(): Entry {
        return this.$store.getters.nowPlaying;
    }

    get storePaused(): boolean {
        return this.$store.state.paused;
    }

    get storeVolume(): number {
        return this.$store.getters.volume;
    }

    get audioElement(): HTMLAudioElement {
        return this.$refs.audio as HTMLAudioElement;
    }

    @Watch('storeNowPlaying')
    onStoreNowPlayingChanged(): void {
        if (!this.storeNowPlaying) {
            this.currentNowPlaying = null;
            this.tearDownStream();
            // this.audioElement.pau
            return;
        }

        if (!this.currentNowPlaying || this.currentNowPlaying !== this.storeNowPlaying) {
            this.currentNowPlaying = this.storeNowPlaying;
            this.startStream(this.storeNowPlaying.track);
        }
    }

    @Watch('storePaused')
    onPausedChanged(val: boolean): void {
        if (val) {
            this.audioElement.pause();
        } else {
            this.audioElement.play();
        }
    }

    @Watch('storeVolume')
    onStoreVolumeChanged(volume: number): void {
        this.audioElement.volume = volume;
    }

    created(): void {
        this.intervalID = window.setInterval(this.emitValues, 200);
    }

    mounted(): void {
        this.audioElement.volume = this.storeVolume;
        this.$root.$on(seekEvent, (position: number) => {
            if (!this.storeNowPlaying) {
                return;
            }
            const trackTarget = this.storeNowPlaying.track.duration * position;
            const localTarget = trackTarget - this.streamStartOffset;
            const loaded = this.audioElement.duration;
            if (localTarget >= 0 && isFinite(loaded) && localTarget < loaded) {
                this.audioElement.currentTime = localTarget;
            } else {
                this.startStream(this.storeNowPlaying.track, trackTarget);
            }
        });
    }

    destroyed(): void {
        window.clearInterval(this.intervalID);
        this.tearDownStream();
    }

    onAudioElementEnded(): void {
        this.currentNowPlaying = null;
        this.$store.commit(Mutation.Next);
    }

    onAudioElementError(): void {
        Notifications.pushError(this, `Could not play "${this.storeNowPlaying.track.title}".`);
        this.$store.commit(Mutation.Next);
    }

    onAudioElementPlay(): void {
        this.$store.commit(Mutation.Play);
    }

    onAudioElementPause(): void {
        this.$store.commit(Mutation.Pause);
    }

    private startStream(track: Track, seekSeconds?: number): void {
        this.tearDownStream();
        if (!Hls.isSupported()) {
            Notifications.pushError(this, 'Your browser does not support HLS playback.');
            return;
        }

        this.streamStartOffset = seekSeconds && seekSeconds > 0 ? seekSeconds : 0;

        const ws = new WebSocket(this.apiService.streamWebSocketUrl(track, seekSeconds));
        this.streamWs = ws;

        ws.onmessage = (event) => {
            if (this.streamWs !== ws) {
                return;
            }
            const streamId = typeof event.data === 'string' ? event.data : '';
            if (!streamId) {
                Notifications.pushError(this, `Could not start streaming "${track.title}".`);
                return;
            }
            this.loadHls(this.apiService.streamPlaylistUrl(track, streamId));
            this.audioElement.play();
        };

        ws.onerror = () => {
            if (this.streamWs !== ws) {
                return;
            }
            Notifications.pushError(this, `Could not start streaming "${track.title}".`);
        };
    }

    private loadHls(url: string): void {
        this.destroyHls();
        this.hls = new Hls({
            xhrSetup: (xhr) => {
                xhr.withCredentials = true;
            },
            startPosition: 0,
        });
        this.hls.on(Hls.Events.ERROR, (_, data) => {
            console.error('hls.js error', data);
        });
        this.hls.loadSource(url);
        this.hls.attachMedia(this.audioElement);
    }

    private destroyHls(): void {
        if (this.hls) {
            this.hls.destroy();
            this.hls = null;
        }
    }

    private closeStreamWs(): void {
        if (this.streamWs) {
            this.streamWs.onmessage = null;
            this.streamWs.onerror = null;
            this.streamWs.onclose = null;
            this.streamWs.close();
            this.streamWs = null;
        }
    }

    private tearDownStream(): void {
        this.destroyHls();
        this.closeStreamWs();
    }

    private emitValues(): void {
        if (this.audioElement) {
            const playbackData: PlaybackData = {
                currentTime: this.streamStartOffset + this.audioElement.currentTime,
            };
            this.$emit('playback-data', playbackData);
        }
    }
}
