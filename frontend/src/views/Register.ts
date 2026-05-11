import { Component, Vue, Watch } from 'vue-property-decorator';
import MainHeader from '@/components/MainHeader.vue';
import SubHeader from '@/components/SubHeader.vue';
import FormInput from '@/components/forms/FormInput.vue';
import AppButton from '@/components/forms/AppButton.vue';
import { ApiService } from '@/services/ApiService';
import Notifications from '@/components/Notifications';
import { LoginCommand } from '@/dto/LoginCommand';
import ActionBar from '@/components/ActionBar.vue';
import ActionBarButton from '@/components/ActionBarButton.vue';
import { RegisterCommand } from '@/dto/RegisterCommand';
import { User } from '@/dto/User';


@Component({
    components: {
        MainHeader,
        SubHeader,
        FormInput,
        AppButton,
        ActionBar,
        ActionBarButton,
    },
})
export default class Register extends Vue {

    cmd = new RegisterCommand();
    working = false;

    private readonly apiService = new ApiService(this);
    private readonly errLogin = `Sign up succeeded but the automatic login failed.`;
    private readonly errSignUp = `Error during the sign up process.`;

    @Watch('user', {immediate: true})
    onUserChanged(user: User): void {
        if (user) {
            this.escapeToBrowse();
        }
    }

    submit(): void {
        if (!this.formValid) {
            return;
        }

        this.working = true;
        this.cmd.token = this.$route.params.token;
        this.apiService.register(this.cmd)
            .then(
                () => {
                    const loginCommand: LoginCommand = {username: this.cmd.username, password: this.cmd.password};
                    this.apiService.login(loginCommand)
                        .then(
                            () => {
                                this.working = false;
                                this.escapeToBrowse();
                            },
                            error => {
                                this.working = false;
                                Notifications.pushError(this, this.errLogin, error);
                            },
                        );
                },
                error => {
                    this.working = false;
                    Notifications.pushError(this, this.errSignUp, error);
                },
            );
    }

    escapeToBrowse(): void {
        this.$router.replace({name: 'browse'});
    }

    goToBrowse(): void {
        this.$router.push({name: 'browse'});
    }

    get formValid(): boolean {
        return !this.formErrors || this.formErrors.length === 0;
    }

    get formErrors(): string[] {
        const errors = [];

        if (!this.cmd.username) {
            errors.push('Username can not be empty.');
        }

        if (!this.cmd.password) {
            errors.push('Password can not be empty.');
        }

        return errors;
    }
}
