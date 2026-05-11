<template>
    <div class="invitations">
        <app-button class="button" text="Create an invitation link" :working="working" @click="createInvitation"
                    v-tooltip="'Create a link containing a one-time token allowing your friend to register an account.'"></app-button>
        <pre class="token" v-if="invitationUrl" @click="copy" v-tooltip="'Click to copy.'">{{ invitationUrl }}</pre>
    </div>
</template>
<script lang="ts" src="./Invitations.ts"></script>
<style scoped lang="scss" src="./Invitations.scss"></style>
