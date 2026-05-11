import { Component, Prop, Vue, Ref } from 'vue-property-decorator';
import { Track } from '@/dto/Track';
import { Entry } from '@/dto/Entry';
import { Mutation, ReplaceCommand, AppendCommand, RemoveCommand } from '@/store';
import { TextService } from '@/services/TextService';
import Notifications from '@/components/Notifications';

import Spinner from '@/components/Spinner.vue';
import Dropdown from '@/components/Dropdown.vue';
import DropdownElement from '@/components/DropdownElement.vue';
import DropdownDivider from '@/components/DropdownDivider.vue';


@Component({
    components: {
        Spinner,
        Dropdown,
        DropdownElement,
        DropdownDivider,
    },
})
export default class Tracks extends Vue {

    @Prop()
    entries: Entry[];

    @Prop()
    queueMode: boolean;

    @Prop()
    showGoToAlbum: boolean;

    @Ref('dropdowns')
    dropdowns: Dropdown[];

    private readonly textService = new TextService();

    isNowPlaying(index: number, track: Track): boolean {
        const nowPlaying: Entry = this.$store.getters.nowPlaying;
        if (nowPlaying) {
            if (this.queueMode) {
                return index === this.$store.state.playingIndex;
            } else {
                return track.id === nowPlaying.track.id;
            }
        }
        return false;
    }

    showPlayAlbumButtonAsPause(index: number, track: Track): boolean {
        if (!this.nowPlaying) {
            return false;
        }

        if (this.queueMode) {
            if (index === this.$store.state.playingIndex) {
                return !this.paused();
            }
        } else {
            if (this.nowPlaying.track.id === track.id) {
                return !this.paused();
            }
        }

        return false;
    }

    isBeingConverted(track: Track): boolean {
        return !track.duration;
    }

    onPlayTrackButtonClicked(index: number): void {
        if (this.isNowPlaying(index, this.entries[index].track)) {
            if (this.paused()) {
                this.$store.commit(Mutation.Play);
            } else {
                this.$store.commit(Mutation.Pause);
            }
        } else {
            // I am very tired but something tells me it may be a good idea to copy
            // this instead of using it directly
            const entries: Entry[] = this.entries
                .map(v => {
                    return {
                        album: v.album,
                        track: v.track,
                    };
                });
            const command: ReplaceCommand = {
                entries: entries,
                playingIndex: index,
            };
            this.$store.commit(Mutation.Replace, command);
        }
    }

    addToQueue(entry: Entry): void {
        const command: AppendCommand = {
            entries: [entry],
        };
        this.$store.commit(Mutation.Append, command);
        this.closeDropdowns();
        Notifications.pushSuccess(this, 'Song added to queue.');
    }

    removeFromQueue(entry: Entry): void {
        const command: RemoveCommand = {
            entry: entry,
        };
        this.$store.commit(Mutation.Remove, command);
        this.closeDropdowns();
    }

    goToAlbum(entry: Entry): void {
        this.closeDropdowns();
        this.$emit('select-album', entry.album);
    }

    formatDuration(track: Track): string {
        return this.textService.formatTime(track.duration);
    }

    get nowPlaying(): Entry {
        return this.$store.getters.nowPlaying;
    }

    private paused(): boolean {
        return this.$store.getters.nowPlayingPaused;
    }

    private closeDropdowns(): void {
        for (const dropdown of this.dropdowns) {
            dropdown.close();
        }
    }

}
