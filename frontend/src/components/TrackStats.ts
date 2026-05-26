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
        return this.humanize(this.stats.streamCacheSize);
    }

    private humanize(bytes: number): string {
        const options = {
            round: 2,
            base: 10,
        };
        return filesize(bytes, options);
    }

}
