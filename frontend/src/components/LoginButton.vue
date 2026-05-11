<template>
    <div class="login-button">
        <div class="button" v-if="loading">
            <spinner></spinner>
        </div>
        <a class="button" v-if="!loading && user" v-tooltip="'You are signed in as '+ user.username + '.'" @click="settings">
            <i class="fas fa-user-check"></i>
        </a>
        <a class="button" v-if="!loading && !user" v-tooltip="'You are not signed in, some content may not be visible.'" @click="login">
            <i class="fas fa-user-times"></i>
        </a>
    </div>
</template>
<script lang="ts" src="./LoginButton.ts"></script>
<style scoped lang="scss" src="./LoginButton.scss"></style>
