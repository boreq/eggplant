import { Component, Prop, Vue } from 'vue-property-decorator';


@Component
export default class SubHeader extends Vue {

    @Prop()
    text: string;

}
