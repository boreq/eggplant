import { Component, Prop, Vue } from 'vue-property-decorator';

@Component
export default class FormInput extends Vue {

    @Prop()
    type: string;

    @Prop()
    placeholder: string;

    @Prop()
    value: any;

    @Prop()
    icon: string;

    onInput(event: Event): void {
        this.$emit('input', (event.target as HTMLInputElement).value);
    }

    onKeyDown(event: KeyboardEvent): void {
        if (event.key === 'Enter') {
            this.$emit('submit');
        }
    }

}
