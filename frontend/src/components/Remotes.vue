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
                <div class="address">
                    {{ remote.address }}
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
            There are no connected instances yet.
        </p>
    </div>
</template>
<script lang="ts" src="./Remotes.ts"></script>
<style scoped lang="scss" src="./Remotes.scss"></style>
