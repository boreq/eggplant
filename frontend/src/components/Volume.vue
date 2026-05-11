<template>
    <div class="volume">
        <a class="button" @click="toggleMute()">
            <i class="fas fa-volume-up" v-if="!muted"></i>
            <i class="fas fa-volume-mute" v-if="muted"></i>
        </a>
        <div class="bar">
            <VerticalProgressBar :value="volume" @value-selected="changeVolume"></VerticalProgressBar>
        </div>
    </div>
</template>
<script lang="ts" src="./Volume.ts"></script>
<style scoped lang="scss" src="./Volume.scss"></style>
