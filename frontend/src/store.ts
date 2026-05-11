import Vue from 'vue';
import Vuex from 'vuex';
import VuexPersistence from 'vuex-persist';
import { Entry } from '@/dto/Entry';
import { User } from '@/dto/User';

Vue.use(Vuex);

const vuexPersist = new VuexPersistence<State>({
    storage: window.localStorage,
    reducer: (state) => ({
        volume: state.volume,
        muted: state.muted,
        shuffle: state.shuffle,
    }),
});

export enum Mutation {
    Replace = 'replace',
    Append = 'append',
    Remove = 'remove',
    Play = 'play',
    Pause = 'pause',
    Previous = 'previous',
    Next = 'next',
    SetVolume = 'setVolume',
    Mute = 'mute',
    Unmute = 'unmute',
    Shuffle = 'shuffle',
    NoShuffle = 'noShuffle',
    SetUser = 'setUser',
}


export class State {
    entries: Entry[];
    playingIndex: number;
    paused: boolean;
    volume: number; // [0, 1]
    muted: boolean;
    shuffle: boolean;
    user: User;
}

export class ReplaceCommand {
    entries: Entry[];
    playingIndex: number;
}

export class AppendCommand {
    entries: Entry[];
}

export class RemoveCommand {
    entry: Entry;
}

export class SetVolumeCommand {
    volume: number;
}

export default new Vuex.Store<State>({
    state: {
        entries: [],
        playingIndex: null,
        paused: true,
        volume: 0.5,
        muted: false,
        shuffle: false,
        user: undefined,
    },
    mutations: {
        [Mutation.Replace](state: State, command: ReplaceCommand): void {
            state.entries = command.entries;

            if (command.playingIndex < 0) {
                state.playingIndex = 0;
            } else if (command.playingIndex > command.entries.length - 1) {
                state.playingIndex = command.entries.length - 1;
            } else {
                state.playingIndex = command.playingIndex;
            }
        },
        [Mutation.Append](state: State, command: AppendCommand): void {
            state.entries = state.entries.concat(command.entries);

            if (state.playingIndex === undefined || state.playingIndex === null) {
                state.playingIndex = 0;
            }
        },
        [Mutation.Remove](state: State, command: RemoveCommand): void {
            const index = state.entries.indexOf(command.entry);
            if (index >= 0) {
                state.entries.splice(index, 1);

                if (state.playingIndex > index) {
                    state.playingIndex -= 1;
                }
            }

            if (state.entries.length > 0) {
                if (state.playingIndex > state.entries.length - 1) {
                    state.playingIndex = state.entries.length - 1;
                }
            } else {
                state.playingIndex = undefined;
            }
        },
        [Mutation.Play](state: State): void {
            if (!emptyArray(state.entries)) {
                state.paused = false;
            }
        },
        [Mutation.Pause](state: State): void {
            state.paused = true;
        },
        [Mutation.Previous](state: State): void {
            if (!emptyArray(state.entries)) {
                if (state.shuffle && state.entries.length > 1) {
                    const previousIndex = state.playingIndex;
                    while (state.playingIndex === previousIndex) {
                        state.playingIndex = getRandomInt(0, state.entries.length);
                    }
                } else {
                    state.playingIndex -= 1;
                    if (state.playingIndex < 0) {
                        state.playingIndex = state.entries.length - 1;
                    }
                }
            }
        },
        [Mutation.Next](state: State): void {
            if (!emptyArray(state.entries)) {
                if (state.shuffle && state.entries.length > 1) {
                    const previousIndex = state.playingIndex;
                    while (state.playingIndex === previousIndex) {
                        state.playingIndex = getRandomInt(0, state.entries.length);
                    }
                } else {
                    state.playingIndex += 1;
                    if (state.playingIndex > state.entries.length - 1) {
                        state.playingIndex = 0;
                    }
                }
            }
        },
        [Mutation.SetVolume](state: State, command: SetVolumeCommand): void {
            state.volume = command.volume;
            if (state.volume < 0) {
                state.volume = 0;
            }
            if (state.volume > 1) {
                state.volume = 1;
            }
            state.muted = state.volume === 0;
        },
        [Mutation.Mute](state: State): void {
            state.muted = true;
        },
        [Mutation.Unmute](state: State): void {
            state.muted = false;
        },
        [Mutation.Shuffle](state: State): void {
            state.shuffle = true;
        },
        [Mutation.NoShuffle](state: State): void {
            state.shuffle = false;
        },
        [Mutation.SetUser](state: State, user: User): void {
            state.user = user;
        },
    },
    getters: {
        nowPlaying: (state: State): Entry => {
            return nowPlaying(state);
        },
        nowPlayingPaused: (state: State): boolean => {
            return state.paused || !nowPlaying(state);
        },
        volume: (state: State): number => {
            if (state.muted) {
                return 0;
            }
            return state.volume;
        },
    },
    plugins: [vuexPersist.plugin],
});

function nowPlaying(state: State): Entry {
    if (emptyArray(state.entries)) {
        return null;
    }
    return state.entries[state.playingIndex];
}

function emptyArray(arr: any[]): boolean {
    return !arr || arr.length === 0;
}

function getRandomInt(min, max) {
    min = Math.ceil(min);
    max = Math.floor(max);
    return Math.floor(Math.random() * (max - min)) + min; // The maximum is exclusive and the minimum is inclusive
}
