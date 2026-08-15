import { Component, Prop, Vue } from 'vue-property-decorator';
import { RemoteLibrary } from '@/dto/RemoteLibrary';


@Component
export default class RemoteLibraries extends Vue {

    @Prop()
    libraries: RemoteLibrary[];

    selectLibrary(library: RemoteLibrary): void {
        this.$emit('select-library', library);
    }

}
