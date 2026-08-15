import { Album } from '@/dto/Album';
import axios, { AxiosResponse } from 'axios'; // do not add { }, some webshit bs?
import { Track } from '@/dto/Track';
import { Stats } from '@/dto/Stats';
import { Thumbnail } from '@/dto/Thumbnail';
import { Version } from '@/dto/Version';
import { CommandInitialize } from '@/dto/CommandInitialize';
import { LoginCommand } from '@/dto/LoginCommand';
import { LoginResponse } from '@/dto/LoginResponse';
import { Mutation } from '@/store';
import { AuthService } from '@/services/AuthService';
import { User } from '@/dto/User';
import { Invitation } from '@/dto/Invitation';
import { RegisterCommand } from '@/dto/RegisterCommand';
import { StreamStartResponse } from '@/dto/StreamStartResponse';
import { TrackDuration } from '@/dto/TrackDuration';
import { RemoteInstance } from '@/dto/RemoteInstance';
import { Library } from '@/dto/Library';
import { AddRemoteResult } from '@/dto/AddRemoteResult';

/*
declare module 'vue-property-decorator' {
    interface Vue {
        $cookie: any;
    }
}
*/

export class ApiService {

    private readonly axios = axios.create({withCredentials: import.meta.env.DEV});
    private readonly authService = new AuthService();

    constructor(private vue: any) {
        this.axios.interceptors.response.use(
            response => {
                return response;
            },
            error => {
                if (error.response && error.response.status === 401) {
                    this.authService.clearToken();
                    this.vue.$store.commit(Mutation.SetUser, null);
                }
                return Promise.reject(error);
            });
    }

    browse(id?: string, instanceId?: string): Promise<AxiosResponse<Album>> {
        let url: string;
        if (instanceId) {
            const instance = `remote/${encodeURIComponent(instanceId)}`;
            url = id ? `${instance}/browse/${id}` : `${instance}/browse`;
        } else {
            url = id ? `browse/${id}` : 'browse';
        }
        return this.axios.get<Album>(import.meta.env.VUE_APP_API_PREFIX + url);
    }

    search(query: string): Promise<AxiosResponse<any>> {
        const url = `search`;
        return this.axios.get<any>(
            import.meta.env.VUE_APP_API_PREFIX + url,
            {
                params: {
                    query: query,
                },
            },
        );
    }

    stats(): Promise<AxiosResponse<Stats>> {
        const url = `stats`;
        return this.axios.get<Stats>(import.meta.env.VUE_APP_API_PREFIX + url);
    }

    // trackBase returns the API path prefix for a track. Remote tracks are
    // proxied through this instance under the remote-instance prefix; local
    // tracks use the plain track path.
    private trackBase(track: Track): string {
        if (track.remoteInstanceId) {
            return `remote/${encodeURIComponent(track.remoteInstanceId)}/track/${track.id}`;
        }
        return `track/${track.id}`;
    }

    getTrackDuration(track: Track): Promise<AxiosResponse<TrackDuration>> {
        const url = `${this.trackBase(track)}/duration`;
        return this.axios.get<TrackDuration>(import.meta.env.VUE_APP_API_PREFIX + url);
    }

    startStream(track: Track, seekSeconds?: number): Promise<AxiosResponse<StreamStartResponse>> {
        let url = `${this.trackBase(track)}/stream`;
        if (seekSeconds !== undefined && seekSeconds > 0) {
            url += `?seek=${encodeURIComponent(seekSeconds.toString())}`;
        }
        return this.axios.post<StreamStartResponse>(import.meta.env.VUE_APP_API_PREFIX + url);
    }

    keepStreamAlive(track: Track, streamId: string): Promise<void> {
        const url = `${this.trackBase(track)}/stream/${streamId}/keepalive`;
        return this.axios.post(import.meta.env.VUE_APP_API_PREFIX + url).then(() => {});
    }

    streamPlaylistUrl(track: Track, streamId: string): string {
        const url = `${this.trackBase(track)}/stream/${streamId}/playlist`;
        return import.meta.env.VUE_APP_API_PREFIX + url;
    }

    thumbnailUrl(thumbnail: Thumbnail, instanceId?: string): string {
        const url = instanceId
            ? `remote/${encodeURIComponent(instanceId)}/thumbnail/${thumbnail.id}`
            : `thumbnail/${thumbnail.id}`;
        return import.meta.env.VUE_APP_API_PREFIX + url;
    }

    initialize(cmd: CommandInitialize): Promise<AxiosResponse<void>> {
        const url = `auth/register-initial`;
        return this.axios.post<void>(import.meta.env.VUE_APP_API_PREFIX + url, cmd);
    }

    login(cmd: LoginCommand): Promise<void> {
        const url = `auth/login`;
        return new Promise((resolve, reject) => {
            this.axios.post<LoginResponse>(import.meta.env.VUE_APP_API_PREFIX + url, cmd)
                .then(
                    response => {
                        this.authService.storeToken(response.data.token);
                        this.refreshCurrentUser()
                            .then(
                                () => {
                                    resolve();
                                },
                                error => {
                                    reject(error);
                                },
                            );
                    },
                    error => {
                        reject(error);
                    },
                );
        });
    }

    logout(): Promise<void> {
        const url = `auth/logout`;
        return new Promise((resolve, reject) => {
            this.axios.post<void>(import.meta.env.VUE_APP_API_PREFIX + url)
                .then(
                    () => {
                        this.authService.clearToken();
                        this.vue.$store.commit(Mutation.SetUser, null);
                        resolve();
                    },
                    error => {
                        reject(error);
                    },
                );
        });
    }

    refreshCurrentUser(): Promise<User> {
        const url = `auth`;
        return new Promise((resolve, reject) => {
            this.axios.get<User>(import.meta.env.VUE_APP_API_PREFIX + url)
                .then(
                    response => {
                        this.vue.$store.commit(Mutation.SetUser, response.data);
                        resolve(response.data);
                    },
                    error => {
                        if (error.response && error.response.status === 401) {
                            resolve(null);
                        } else {
                            reject(error);
                        }
                    },
                );
        });
    }

    getVersion(): Promise<AxiosResponse<Version>> {
        const url = `version`;
        return this.axios.get<Version>(import.meta.env.VUE_APP_API_PREFIX + url);
    }

    register(cmd: RegisterCommand): Promise<AxiosResponse<void>> {
        const url = `auth/register`;
        return this.axios.post<void>(import.meta.env.VUE_APP_API_PREFIX + url, cmd);
    }

    list(): Promise<AxiosResponse<User[]>> {
        const url = `auth/users`;
        return this.axios.get<User[]>(import.meta.env.VUE_APP_API_PREFIX + url);
    }

    createInvitation(): Promise<AxiosResponse<Invitation>> {
        const url = `auth/create-invitation`;
        return this.axios.post<Invitation>(import.meta.env.VUE_APP_API_PREFIX + url);
    }

    remove(username: string): Promise<AxiosResponse<void>> {
        username = encodeURIComponent(username);
        const url = `auth/users/${username}`;
        return this.axios.delete<void>(import.meta.env.VUE_APP_API_PREFIX + url);
    }

    listLibraries(): Promise<AxiosResponse<Library[]>> {
        const url = `library`;
        return this.axios.get<Library[]>(import.meta.env.VUE_APP_API_PREFIX + url);
    }

    listRemotes(): Promise<AxiosResponse<RemoteInstance[]>> {
        const url = `remote`;
        return this.axios.get<RemoteInstance[]>(import.meta.env.VUE_APP_API_PREFIX + url);
    }

    addRemote(remoteUrl: string): Promise<AxiosResponse<AddRemoteResult>> {
        const url = `remote`;
        return this.axios.post<AddRemoteResult>(import.meta.env.VUE_APP_API_PREFIX + url, { url: remoteUrl });
    }

    setRemotePairingToken(id: string, peerToken: string): Promise<AxiosResponse<void>> {
        const url = `remote/${encodeURIComponent(id)}/pairing-token`;
        return this.axios.post<void>(import.meta.env.VUE_APP_API_PREFIX + url, { peer_token: peerToken });
    }

}
