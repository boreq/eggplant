import { Component, Prop, Vue } from 'vue-property-decorator';
import { ThumbnailStats as ThumbnailStatsDto } from '@/dto/Stats';
import filesize from 'filesize';


@Component
export default class ThumbnailStats extends Vue {

    @Prop()
    stats: ThumbnailStatsDto;

    get originalSize(): string {
        return this.humanize(this.stats.originalSize);
    }

    get convertedSize(): string {
        return this.humanize(this.stats.convertedSize);
    }

    private humanize(bytes: number): string {
        const options = {
            round: 2,
            base: 10,
        };
        return filesize(bytes, options);
    }

}
