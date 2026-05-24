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

    get nowPlaying(): Entry {
        return this.$store.getters.nowPlaying;
    }

    get paused(): boolean {
        return this.$store.state.paused;
    }

    get audio(): HTMLAudioElement {
        return this.$refs.audio as HTMLAudioElement;
    }

    get volume(): number {
        return this.$store.getters.volume;
    }

    @Watch('nowPlaying')
    onNowPlayingChanged(): void {
        if (!this.nowPlaying) {
            this.currentNowPlaying = null;
            this.tearDownStream();
            this.pause();
            return;
        }

        if (!this.currentNowPlaying || this.currentNowPlaying !== this.nowPlaying) {
            this.currentNowPlaying = this.nowPlaying;
            this.startStream(this.nowPlaying.track);
        }
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
            this.play();
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
        this.hls.attachMedia(this.audio);
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

    @Watch('paused')
    onPausedChanged(val: boolean): void {
        if (val) {
            this.audio.pause();
        } else {
            this.audio.play();
        }
    }

    @Watch('volume')
    onVolumeChanged(volume: number): void {
        this.audio.volume = volume;
    }

    created(): void {
        this.intervalID = window.setInterval(this.emitValues, 100);
    }

    mounted(): void {
        this.audio.volume = this.volume;
        this.$root.$on(seekEvent, (position: number) => {
            if (!this.nowPlaying) {
                return;
            }
            const trackTarget = this.nowPlaying.track.duration * position;
            const localTarget = trackTarget - this.streamStartOffset;
            const loaded = this.audio.duration;
            if (localTarget >= 0 && isFinite(loaded) && localTarget < loaded) {
                this.audio.currentTime = localTarget;
            } else {
                this.startStream(this.nowPlaying.track, trackTarget);
            }
        });
    }

    destroyed(): void {
        window.clearInterval(this.intervalID);
        this.tearDownStream();
    }

    onEnded(): void {
        this.currentNowPlaying = null;
        this.$store.commit(Mutation.Next);
    }

    onError(): void {
        Notifications.pushError(this, `Could not play "${this.nowPlaying.track.title}".`);
        this.$store.commit(Mutation.Next);
    }

    onPlay(): void {
        this.$store.commit(Mutation.Play);
    }

    onPause(): void {
        this.$store.commit(Mutation.Pause);
    }

    private emitValues(): void {
        if (this.audio) {
            const playbackData: PlaybackData = {
                currentTime: this.streamStartOffset + this.audio.currentTime,
            };
            this.$emit('playback-data', playbackData);
        }
    }

    private play(): void {
        this.$store.commit(Mutation.Play);
        this.audio.play();
    }

    private pause(): void {
        this.$store.commit(Mutation.Pause);
        this.audio.pause();
    }

}
