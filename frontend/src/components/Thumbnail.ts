import { Component, Prop, Ref, Vue } from 'vue-property-decorator';
import { BasicAlbum } from '@/dto/BasicAlbum';
import { ApiService } from '@/services/ApiService';
import Spinner from '@/components/Spinner.vue';


@Component({
    components: {
        Spinner,
    },
})
export default class Thumbnail extends Vue {

    @Prop()
    album: BasicAlbum;

    @Prop({ type: Boolean, default: false })
    tilt: boolean;

    @Ref()
    image: HTMLImageElement;

    @Ref()
    root: HTMLElement;

    converting = false;

    private timeoutId: number;

    private readonly apiService = new ApiService(this);

    private readonly maxTilt = 15;

    get thumbnailUrl(): string {
        if (this.album) {
            return this.apiService.thumbnailUrl(this.album.thumbnail);
        }
        return null;
    }

    destroyed(): void {
        this.clearTimeout();
    }

    onError(): void {
        this.converting = true;
        this.reload();
    }

    onLoad(): void {
        this.converting = false;
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

    private reload(): void {
        this.clearTimeout();
        this.timeoutId = window.setTimeout(() => this.image.src = this.thumbnailUrl, 5000);
    }

    private clearTimeout(): void {
        if (this.timeoutId) {
            window.clearTimeout(this.timeoutId);
        }
    }

}
