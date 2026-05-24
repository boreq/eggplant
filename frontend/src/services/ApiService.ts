import { Album } from '@/dto/Album';
import axios, { AxiosResponse } from 'axios'; // do not add { }, some webshit bs?
import { Track } from '@/dto/Track';
import { Stats } from '@/dto/Stats';
import { Thumbnail } from '@/dto/Thumbnail';
import { CommandInitialize } from '@/dto/CommandInitialize';
import { LoginCommand } from '@/dto/LoginCommand';
import { LoginResponse } from '@/dto/LoginResponse';
import { Mutation } from '@/store';
import { AuthService } from '@/services/AuthService';
import { User } from '@/dto/User';
import { Invitation } from '@/dto/Invitation';
import { RegisterCommand } from '@/dto/RegisterCommand';

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

    browse(id?: string): Promise<AxiosResponse<Album>> {
        const url = id ? `browse/${id}` : 'browse';
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

    streamWebSocketUrl(track: Track, seekSeconds?: number): string {
        let url = import.meta.env.VUE_APP_WS_PREFIX + `track/${track.id}/stream`;
        if (seekSeconds !== undefined && seekSeconds > 0) {
            url += `?seek=${encodeURIComponent(seekSeconds.toString())}`;
        }
        return url;
    }

    streamPlaylistUrl(track: Track, streamId: string): string {
        const url = `track/${track.id}/stream/${streamId}/playlist`;
        return import.meta.env.VUE_APP_API_PREFIX + url;
    }

    thumbnailUrl(thumbnail: Thumbnail): string {
        const url = `thumbnail/${thumbnail.id}`;
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

    getVersion(): Promise<AxiosResponse<{ version: string }>> {
        const url = `version`;
        return this.axios.get<{ version: string }>(import.meta.env.VUE_APP_API_PREFIX + url);
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
        const url = `auth/users/${username}/remove`;
        return this.axios.post<void>(import.meta.env.VUE_APP_API_PREFIX + url);
    }

}
