<template>
    <div class="register">
        <action-bar>
            <action-bar-button text="Go back to the music" icon="fas fa-arrow-left" @click="goToBrowse"></action-bar-button>
        </action-bar>
        <main-header text="Sign up"></main-header>

        <div class="center">
            <form>
                <form-input type="text" placeholder="Username" icon="fas fa-user"
                            v-model="cmd.username" @submit="submit"></form-input>
                <form-input type="password" placeholder="Password" icon="fas fa-key"
                            v-model="cmd.password" @submit="submit"></form-input>
                <app-button text="Sign up" @click="submit" :disabled="!formValid" :working="working"></app-button>
            </form>
        </div>
    </div>
</template>
<script lang="ts" src="./Register.ts"></script>
<style scoped lang="scss" src="./Register.scss"></style>
