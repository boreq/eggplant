import { Component, Prop, Vue } from 'vue-property-decorator';
import { BasicAlbum } from '@/dto/BasicAlbum';
import Thumbnail from '@/components/Thumbnail.vue';


@Component({
    components: {
        Thumbnail,
    },
})
export default class Albums extends Vue {

    @Prop()
    albums: BasicAlbum[];

    selectAlbum(album: BasicAlbum): void {
        this.$emit('select-album', album);
    }

}
