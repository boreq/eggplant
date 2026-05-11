import { Component, Vue, Watch } from 'vue-property-decorator';
import MainHeader from '@/components/MainHeader.vue';
import SubHeader from '@/components/SubHeader.vue';
import FormInput from '@/components/forms/FormInput.vue';
import AppButton from '@/components/forms/AppButton.vue';
import ActionBarButton from '@/components/ActionBarButton.vue';
import ActionBar from '@/components/ActionBar.vue';
import { ApiService } from '@/services/ApiService';
import Notifications from '@/components/Notifications';
import { User } from '@/dto/User';
import Users from '@/components/Users.vue';
import Invitations from '@/components/Invitations.vue';


@Component({
    components: {
        MainHeader,
        SubHeader,
        FormInput,
        AppButton,
        ActionBar,
        ActionBarButton,
        Users,
        Invitations,
    },
})
export default class Settings extends Vue {

    logoutInProgress = false;
    private readonly apiService = new ApiService(this);

    @Watch('user', {immediate: true})
    onUserChanged(user: User): void {
        if (user === null) {
            this.escapeToBrowse();
        }
    }

    escapeToBrowse(): void {
        this.$router.replace({name: 'browse'});
    }

    goToBrowse(): void {
        this.$router.push({name: 'browse'});
    }

    goToStats(): void {
        this.$router.push({name: 'stats'});
    }

    logout(): void {
        this.logoutInProgress = true;
        this.apiService.logout()
            .then(
                () => {
                    this.escapeToBrowse();
                },
                error => {
                    Notifications.pushError(this, 'There was an error during the sign out process.', error);
                },
            )
            .finally(
                () => this.logoutInProgress = false,
            );
    }

    get user(): User {
        return this.$store.state.user;
    }
}
