export interface RemoteInstance {
    id: string;
    name: string;
    url: string;
}

export interface AddRemoteCommand {
    name: string;
    url: string;
    username: string;
    password: string;
}
