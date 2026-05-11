import { Component, Vue } from 'vue-property-decorator';
import { User } from '@/dto/User';
import Spinner from '@/components/Spinner.vue';


@Component({
    components: {
        Spinner,
    },
})
export default class LoginButton extends Vue {

    get loading(): boolean {
        return this.user === undefined;
    }

    get user(): User {
        return this.$store.state.user;
    }

    login(): void {
        this.$router.push({name: 'login', query: {next: window.location.pathname}});
    }

    settings(): void {
        this.$router.push({name: 'settings'});
    }

}
