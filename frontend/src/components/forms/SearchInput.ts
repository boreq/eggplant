import { Component, Prop, Vue } from 'vue-property-decorator';

@Component
export default class SearchInput extends Vue {

    @Prop()
    value: any;

    get showCloseButton(): boolean {
        return !!this.value;
    }

    onInput(event: Event): void {
        this.$emit('input', (event.target as HTMLInputElement).value);
    }

}
