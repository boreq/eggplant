import { Component, Vue } from 'vue-property-decorator';
import MainHeader from '@/components/MainHeader.vue';
import SubHeader from '@/components/SubHeader.vue';
import FormInput from '@/components/forms/FormInput.vue';
import AppButton from '@/components/forms/AppButton.vue';
import { CommandInitialize } from '@/dto/CommandInitialize';
import { ApiService } from '@/services/ApiService';
import Notifications from '@/components/Notifications';
import { LoginCommand } from '@/dto/LoginCommand';


@Component({
    components: {
        MainHeader,
        SubHeader,
        FormInput,
        AppButton,
    },
})
export default class Setup extends Vue {

    cmd = new CommandInitialize();
    working = false;

    private readonly errLogin = 'Initial setup succeeded but the automatic login failed.';
    private readonly errSetup = 'Error performing the initial setup.';
    private readonly errCheckSetup = `Could not confirm whether this instance's setup process was completed.`;
    private readonly apiService = new ApiService(this);

    mounted(): void {
        this.redirectAwayIfNeeded();
    }

    submit(): void {
        if (!this.formValid) {
            return;
        }

        this.working = true;
        this.apiService.initialize(this.cmd)
            .then(
                () => {
                    const loginCommand: LoginCommand = {username: this.cmd.username, password: this.cmd.password};
                    this.apiService.login(loginCommand)
                        .then(
                            () => {
                                this.working = false;
                                this.$router.replace({name: 'browse'});
                            },
                            error => {
                                this.working = false;
                                Notifications.pushError(this, this.errLogin, error);
                            },
                        );
                },
                error => {
                    this.working = false;
                    Notifications.pushError(this, this.errSetup, error);
                },
            );
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

    private redirectAwayIfNeeded(): void {
        this.apiService.stats()
            .then(
                response => {
                    if (response.data.users !== 0) {
                        this.$router.replace({name: 'browse'});
                    }
                },
                error => {
                    Notifications.pushError(this, this.errCheckSetup, error);
                });
    }
}
