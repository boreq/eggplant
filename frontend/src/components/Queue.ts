import { Component, Vue } from 'vue-property-decorator';
import { TrackWithAlbum } from '@/dto/TrackWithAlbum';
import { PartialAlbum } from '@/dto/Album';

import MainHeader from '@/components/MainHeader.vue';
import FormInput from '@/components/forms/FormInput.vue';
import AppButton from '@/components/forms/AppButton.vue';
import ActionBarButton from '@/components/ActionBarButton.vue';
import ActionBar from '@/components/ActionBar.vue';
import Tracks from '@/components/Tracks.vue';


@Component({
    components: {
        MainHeader,
        FormInput,
        AppButton,
        ActionBar,
        ActionBarButton,
        Tracks,
    },
})
export default class Queue extends Vue {

    get nowPlaying(): TrackWithAlbum {
        return this.$store.getters.nowPlaying;
    }

    get entries(): TrackWithAlbum[] {
        return this.$store.state.entries;
    }

    get empty(): boolean {
        return !this.entries || this.entries.length === 0;
    }

    onSelectAlbum(album: PartialAlbum): void {
        this.$emit('select-album', album);
    }

}
