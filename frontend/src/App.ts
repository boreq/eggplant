import { Component, Vue } from 'vue-property-decorator';
import { PlaybackData } from '@/dto/PlaybackData';
import { Mutation } from '@/store';
import { ApiService } from '@/services/ApiService';

import Controls from '@/components/Controls.vue';
import Player from '@/components/Player.vue';
import Notifications from '@/components/Notifications.vue';
import NowPlaying from '@/components/NowPlaying.vue';
import MediaSession from '@/components/MediaSession.vue';


@Component({
    components: {
        Controls,
        Player,
        Notifications,
        NowPlaying,
        MediaSession,
    },
})
export default class App extends Vue {

    playbackData: PlaybackData = null;

    private readonly errCheckSetup = `Could not confirm whether this instance's setup process was completed.`;
    private readonly errCheckUser = `Could not retrieve the current user.`;
    private readonly apiService = new ApiService(this);

    created() {
        window.addEventListener('keydown', this.onKeyDown);
        this.redirectToSetupIfNeeded();
        this.loadCurrentUser();
    }

    destroyed(): void {
        window.removeEventListener('keydown', this.onKeyDown);
    }

    onKeyDown(event: KeyboardEvent): void {
        if (event.code === 'Space') {
            if (this.$store.state.paused) {
                this.$store.commit(Mutation.Play);
            } else {
                this.$store.commit(Mutation.Pause);
            }
            event.preventDefault();
        }
    }

    onPlaybackData(playbackData: PlaybackData): void {
        this.playbackData = playbackData;
    }

    get isCsd(): boolean {
        const now = new Date();
        const day = now.getDate();
        const month = now.getMonth() + 1;
        return day === 28 && month === 6;
    }

    private redirectToSetupIfNeeded(): void {
        this.apiService.stats()
            .then(
                response => {
                    if (response.data.users === 0) {
                        this.$router.push({name: 'setup'});
                    }
                },
                error => {
                    Notifications.pushError(this, this.errCheckSetup, error);
                });
    }

    private loadCurrentUser() {
        this.apiService.refreshCurrentUser()
            .catch(error => {
                Notifications.pushError(this, this.errCheckUser, error);
            });
    }
}
