import { Component, Prop, Ref, Vue } from 'vue-property-decorator';
import { ApiService } from '@/services/ApiService';
import { PartialAlbum } from "@/dto/Album";

import Spinner from '@/components/Spinner.vue';

@Component({
    components: {
        Spinner,
    },
})
export default class Thumbnail extends Vue {

    @Prop()
    album: PartialAlbum;

    @Prop({ type: Boolean, default: false })
    tilt: boolean;

    @Ref()
    root: HTMLElement;

    loaded = false;
    errored = false;

    private readonly apiService = new ApiService(this);

    private readonly maxTilt = 15;

    get thumbnailUrl(): string {
        if (this.album) {
            return this.apiService.thumbnailUrl(this.album.thumbnail, this.album.remoteLibraryId);
        }
        return null;
    }

    get showImage(): boolean {
        return !!this.album?.thumbnail && !this.errored;
    }

    onError(): void {
        this.errored = true;
        this.loaded = false;
    }

    onLoad(): void {
        this.loaded = true;
    }

    onMouseMove(e: MouseEvent): void {
        if (!this.tilt || !this.root) return;
        const rect = this.root.getBoundingClientRect();
        const x = (e.clientX - rect.left) / rect.width;
        const y = (e.clientY - rect.top) / rect.height;
        const tiltX = (0.5 - y) * 2 * this.maxTilt;
        const tiltY = (x - 0.5) * 2 * this.maxTilt;
        this.root.style.setProperty('--tilt-x', `${tiltX}deg`);
        this.root.style.setProperty('--tilt-y', `${tiltY}deg`);
        this.root.style.setProperty('--glare-x', `${x * 100}%`);
        this.root.style.setProperty('--glare-y', `${y * 100}%`);
        this.root.style.setProperty('--glare-opacity', '0.35');
    }

    onMouseLeave(): void {
        if (!this.tilt || !this.root) return;
        this.root.style.setProperty('--tilt-x', '0deg');
        this.root.style.setProperty('--tilt-y', '0deg');
        this.root.style.setProperty('--glare-opacity', '0');
    }

}
