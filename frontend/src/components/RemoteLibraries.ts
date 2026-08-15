import { Component, Prop, Vue } from 'vue-property-decorator';
import { Library } from '@/dto/Library';


@Component
export default class RemoteLibraries extends Vue {

    @Prop()
    libraries: Library[];

    selectLibrary(library: Library): void {
        this.$emit('select-library', library);
    }

}
