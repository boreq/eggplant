import { Component, Prop, Vue } from 'vue-property-decorator';
import { StoreStats as StoreStatsDto } from '@/dto/Stats';
import filesize from 'filesize';


@Component
export default class StoreStats extends Vue {

    @Prop()
    stats: StoreStatsDto;

    get conversionProgress(): string {
        if (this.stats.allItems === 0) {
            return '100%';
        }
        const ratio = this.stats.convertedItems / this.stats.allItems;
        return Math.round(ratio * 100) + '%';
    }

    get originalSize(): string {
        return this.humanize(this.stats.originalSize);
    }

    get convertedSize(): string {
        return this.humanize(this.stats.convertedSize);
    }

    get conversionRatio(): string {
        if (this.stats.originalSize === 0) {
            return '100%';
        }
        const ratio = this.stats.convertedSize / this.stats.originalSize;
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
