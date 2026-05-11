import { Component, Prop, Vue } from 'vue-property-decorator';


@Component
export default class MainHeader extends Vue {

    @Prop()
    text: string;

}
