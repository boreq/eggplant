import { Component, Prop, Ref, Vue, Watch } from 'vue-property-decorator';
import { PartialAlbum } from "@/dto/Album";
import Thumbnail from '@/components/Thumbnail.vue';


@Component({
    components: {
        Thumbnail,
    },
})
export default class Albums extends Vue {

    @Prop()
    albums: PartialAlbum[];

    @Prop({ default: false })
    collapsed: boolean;

    @Ref('grid')
    readonly grid: HTMLDivElement;

    columns: number = 1;

    private resizeObserver: ResizeObserver = null;

    mounted(): void {
        this.resizeObserver = new ResizeObserver(() => this.updateColumns());
        this.resizeObserver.observe(this.grid);
        this.updateColumns();
    }

    destroyed(): void {
        if (this.resizeObserver) {
            this.resizeObserver.disconnect();
            this.resizeObserver = null;
        }
    }

    @Watch('togglingAvailable', { immediate: true })
    onTogglingAvailableChanged(): void {
        this.$emit('toggling-available', this.togglingAvailable);
    }

    get togglingAvailable(): boolean {
        return !!this.albums && this.albums.length > this.columns;
    }

    get displayedAlbums(): PartialAlbum[] {
        if (!this.albums) {
            return null;
        }
        if (this.collapsed) {
            return this.albums.slice(0, this.columns);
        }
        return this.albums;
    }

    selectAlbum(album: PartialAlbum): void {
        this.$emit('select-album', album);
    }

    private updateColumns(): void {
        const columns = getComputedStyle(this.grid)
            .gridTemplateColumns
            .split(' ')
            .filter(track => track.endsWith('px'))
            .length;

        if (columns > 0) {
            this.columns = columns;
        }
    }

}
