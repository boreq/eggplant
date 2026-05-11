import { Component, Vue, Watch } from 'vue-property-decorator';
import { Mutation } from '@/store';
import { ApiService } from '@/services/ApiService';
import { PlaybackData } from '@/dto/PlaybackData';
import { Entry } from '@/dto/Entry';
import Notifications from '@/components/Notifications';

export const seekEvent = 'seek';

@Component({
    components: {},
})
export default class Player extends Vue {

    private intervalID: number;

    private readonly apiService = new ApiService(this);

    private currentNowPlaying: Entry = null;

    get nowPlaying(): Entry {
        return this.$store.getters.nowPlaying;
    }

    get paused(): boolean {
        return this.$store.state.paused;
    }

    get nowPlayingUrl(): string {
        const entry = this.nowPlaying;
        if (entry) {
            return this.apiService.trackUrl(entry.track);
        }
        return null;
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
            this.pause();
            return;
        }

        if (!this.currentNowPlaying || this.currentNowPlaying !== this.nowPlaying) {
            this.currentNowPlaying = this.nowPlaying;
            this.audio.src = this.nowPlayingUrl;
            this.play();
        }
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
            if (this.nowPlaying) {
                this.audio.currentTime = this.audio.duration * position;
            }
        });
    }

    destroyed(): void {
        window.clearInterval(this.intervalID);
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
                currentTime: this.audio.currentTime,
                duration: this.audio.duration,
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
