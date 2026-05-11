import { Component, Prop, Vue } from 'vue-property-decorator';


@Component
export default class Indicator extends Vue {

    @Prop({default: false})
    state: boolean;

    @Prop({default: 0})
    width: number;

    private readonly baseWidth = 4;

    get barWidth(): string {
        return `${this.baseWidth + this.width}px`;
    }

}
