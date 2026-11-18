<template>
    <div class="remotes">
        <div class="add">
            <form-input class="input" placeholder="https://eggplant.example.com" :value="newRemoteUrl"
                        @input="v => newRemoteUrl = v" @submit="addRemote"></form-input>
            <app-button class="button" text="Connect instance" :working="adding" @click="addRemote"
                        v-tooltip="'Begin the process of connecting with another Eggplant instance to create a common pool of music with your friends.'"></app-button>
        </div>

        <ul class="list" v-if="remotes && remotes.length">
            <li class="remote" v-for="remote in remotes" :key="remote.id">
                <div class="header">
                    <div class="address">
                        {{ remote.address }}
                    </div>

                    <v-popover offset="16">
                        <a class="disconnect" v-tooltip="'Disconnect this instance.'">
                            <i class="fas fa-times"></i>
                        </a>

                        <template slot="popover">
                            <p>
                                To connect again, you and your friend must pair the instances again.
                            </p>
                            <app-button text="Disconnect" v-close-popover @click="removeRemote(remote)"></app-button>
                        </template>
                    </v-popover>
                </div>

                <div class="details">
                    <div v-if="localPairingToken(remote) && remote.status === 'PAIRING'" class="local-token">
                        <span class="label">Give this token to your friend:</span>
                        <pre class="token" @click="copy(localPairingToken(remote))" v-tooltip="'Click to copy.'">{{ localPairingToken(remote) }}</pre>
                    </div>

                    <div v-if="remote.status === 'PAIRING' && !remote.remote_pairing_token_set" class="peer-token">
                        <span class="label">Paste the token your friend gave you:</span>
                        <div class="row">
                            <form-input class="input" placeholder="Token"
                                        :value="peerTokens[remote.id]"
                                        @input="v => onPeerTokenInput(remote, v)"
                                        @submit="submitPeerToken(remote)"></form-input>
                            <app-button class="button" text="Submit" @click="submitPeerToken(remote)"></app-button>
                        </div>
                    </div>

                    <div class="status" :class="statusClass(remote)">
                        {{ statusText(remote) }}
                    </div>
                </div>
            </li>
        </ul>

        <spinner v-if="!remotes"></spinner>
        <p v-else-if="!remotes.length" class="empty">
            There are no instances yet.
        </p>
    </div>
</template>
<script lang="ts" src="./Remotes.ts"></script>
<style scoped lang="scss" src="./Remotes.scss"></style>
