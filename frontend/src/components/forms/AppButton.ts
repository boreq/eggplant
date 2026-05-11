import { Component, Prop, Vue } from 'vue-property-decorator';
import Spinner from '@/components/Spinner.vue';

@Component({
    components: {
        Spinner,
    },
})
export default class AppButton extends Vue {

    @Prop()
    text: string;

    @Prop()
    disabled: boolean;

    @Prop()
    working: boolean;

    onClick(event: MouseEvent): void {
        if (!this.disabled && !this.working) {
            this.$emit('click', event);
        }
    }

}
