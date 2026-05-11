import { Component, Vue } from 'vue-property-decorator';
import { ApiService } from '@/services/ApiService';
import { User } from '@/dto/User';
import Notifications from '@/components/Notifications';
import AppButton from '@/components/forms/AppButton.vue';
import Spinner from '@/components/Spinner.vue';


@Component({
    components: {
        AppButton,
        Spinner,
    },
})
export default class Users extends Vue {

    users: User[] = null;

    private readonly apiService = new ApiService(this);

    mounted(): void {
        this.load();
    }

    getRole(user: User): string {
        if (user.administrator) {
            return 'Administrator';
        }
        return 'User';
    }

    canRemove(user: User): boolean {
        return user.username !== this.currentUser.username;
    }

    removeUser(user: User): void {
        this.apiService.remove(user.username)
            .then(
                () => {
                    this.load();
                },
                error => {
                    Notifications.pushError(this, `Could not remove the user '${user.username}'.`, error);
                },
            );
    }

    get currentUser(): User {
        return this.$store.state.user;
    }

    private load(): void {
        this.users = null;
        this.apiService.list()
            .then(
                response => {
                    this.users = response.data;
                },
                error => {
                    Notifications.pushError(this, 'Could not list the user accounts.', error);
                },
            );
    }

}
