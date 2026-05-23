import { Component, Prop, Vue } from 'vue-property-decorator';
import { TrackStats as TrackStatsDto } from '@/dto/Stats';
import filesize from 'filesize';


@Component
export default class TrackStats extends Vue {

    @Prop()
    stats: TrackStatsDto;

    get sizeOfTracks(): string {
        return this.humanize(this.stats.sizeOfTracks);
    }

    get sizeOfConvertedTracks(): string {
        return this.humanize(this.stats.sizeOfConvertedTracks);
    }

    get conversionRatio(): string {
        if (this.stats.sizeOfTracks === 0) {
            return '100%';
        }
        const ratio = this.stats.sizeOfConvertedTracks / this.stats.sizeOfTracks;
        return Math.round(ratio * 100) + '%';
    }

    private humanize(bytes: number): string {
        const options = {
            round: 2,
            base: 10,
        };
        return filesize(bytes, options);
    }

}
