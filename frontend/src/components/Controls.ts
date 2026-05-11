import { Component, Prop, Vue } from 'vue-property-decorator';
import { Mutation } from '@/store';
import { Entry } from '@/dto/Entry';
import { PlaybackData } from '@/dto/PlaybackData';
import { TextService } from '@/services/TextService';
import { seekEvent } from '@/components/Player';

import ProgressBar from '@/components/ProgressBar.vue';
import Volume from '@/components/Volume.vue';
import Indicator from '@/components/Indicator.vue';


@Component({
    components: {
        ProgressBar,
        Volume,
        Indicator,
    },
})
export default class Controls extends Vue {

    @Prop()
    playbackData: PlaybackData;

    private textService = new TextService();

    get nowPlaying(): Entry {
        return this.$store.getters.nowPlaying;
    }

    get currentTime(): string {
        if (!this.nowPlaying) {
            return null;
        }

        if (this.playbackData && this.playbackData.currentTime) {
            return this.textService.formatTime(this.playbackData.currentTime);
        }
        return null;
    }

    get duration(): string {
        if (!this.nowPlaying) {
            return null;
        }

        if (this.playbackData && this.playbackData.duration) {
            return this.textService.formatTime(this.playbackData.duration);
        }
        return null;
    }

    get currentTimePercentage(): number {
        if (this.playbackData && this.playbackData.currentTime && this.playbackData.duration) {
            return this.playbackData.currentTime / this.playbackData.duration;
        }
        return 0;
    }

    get paused(): boolean {
        return this.$store.getters.nowPlayingPaused;
    }

    get shuffle(): boolean {
        return this.$store.state.shuffle;
    }

    onPlayPause(): void {
        if (this.paused) {
            this.$store.commit(Mutation.Play);
        } else {
            this.$store.commit(Mutation.Pause);
        }
    }

    onShuffle(): void {
        if (this.shuffle) {
            this.$store.commit(Mutation.NoShuffle);
        } else {
            this.$store.commit(Mutation.Shuffle);
        }
    }

    onPrevious(): void {
        this.$store.commit(Mutation.Previous);
    }

    onNext(): void {
        this.$store.commit(Mutation.Next);
    }

    seek(position: number): void {
        this.$root.$emit(seekEvent, position);
    }

}
