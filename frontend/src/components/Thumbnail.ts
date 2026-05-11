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

    @Ref()
    image: HTMLImageElement;

    converting = false;

    private timeoutId: number;

    private readonly apiService = new ApiService(this);

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
