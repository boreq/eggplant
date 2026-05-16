<template>
    <div class="settings">
        <action-bar>
            <action-bar-button text="Go back to the music" icon="fas fa-arrow-left" @click="goToBrowse"></action-bar-button>
            <action-bar-button text="Sign me out" icon="fas fa-sign-out-alt" @click="logout" :working="logoutInProgress"></action-bar-button>
            <action-bar-button text="Show me the statistics" icon="fas fa-chart-pie" @click="goToStats"></action-bar-button>
        </action-bar>
        <main-header text="Settings"></main-header>

        <spinner v-if="user === undefined"></spinner>

        <template v-else-if="user && user.administrator">
            <sub-header text="Invitations"></sub-header>
            <Invitations></Invitations>

            <sub-header text="Users"></sub-header>
            <Users></Users>
        </template>

        <div v-else class="no-admin-settings">
            There are no additional settings available to your account.
        </div>
    </div>
</template>
<script lang="ts" src="./Settings.ts"></script>
<style scoped lang="scss" src="./Settings.scss"></style>
