<template>
    <div class="player">
        <audio controls ref="audio" @ended="onEnded" @error="onError" @play="onPlay" @pause="onPause"></audio>
    </div>
</template>
<script lang="ts" src="./Player.ts"></script>
<style scoped lang="scss" src="./Player.scss"></style>
