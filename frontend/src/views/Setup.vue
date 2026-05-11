<template>
    <div class="initialize">
        <main-header text="Create an administrator account"></main-header>

        <div class="center">
            <form>
                <form-input type="text" placeholder="Username" icon="fas fa-user"
                            v-model="cmd.username" @submit="submit"></form-input>
                <form-input type="password" placeholder="Password" icon="fas fa-key"
                            v-model="cmd.password" @submit="submit"></form-input>
                <app-button text="Proceed" @click="submit" :disabled="!formValid" :working="working"></app-button>
            </form>
        </div>
    </div>
</template>
<script lang="ts" src="./Setup.ts"></script>
<style scoped lang="scss" src="./Setup.scss"></style>
