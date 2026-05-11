import { Component, Prop, Vue } from 'vue-property-decorator';
import Spinner from '@/components/Spinner.vue';


@Component({
    components: {
        Spinner,
    },
})
export default class ActionBarButton extends Vue {

    @Prop()
    icon: string;

    @Prop()
    text: string;

    @Prop()
    working: boolean;

    onClick(event: MouseEvent): void {
        if (!this.working) {
            this.$emit('click', event);
        }
    }

}
