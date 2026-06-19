export class RemoteInstance {
    id: string;
    address: string;
    status: string;
    remote_pairing_token_set: boolean;
    last_healthcheck_status?: string;
    last_healthcheck_at?: string;
}
