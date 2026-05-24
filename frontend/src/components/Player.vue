<template>
    <div class="player">
        <audio controls ref="audio" @ended="onAudioElementEnded" @error="onAudioElementError" @play="onAudioElementPlay" @pause="onAudioElementPause"></audio>
    </div>
</template>
<script lang="ts" src="./Player.ts"></script>
<style scoped lang="scss" src="./Player.scss"></style>
