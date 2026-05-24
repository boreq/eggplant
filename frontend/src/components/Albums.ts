import { Component, Prop, Vue } from 'vue-property-decorator';
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

    selectAlbum(album: PartialAlbum): void {
        this.$emit('select-album', album);
    }

}
