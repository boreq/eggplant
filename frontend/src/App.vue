<template>
    <div id="app" :class="{csd: isCsd}">
        <div v-if="showPwaWarning" class="pwa-warning">
            To avoid playback issues on mobile install Eggplant as a PWA using Chrome. Once you do you need to disable battery optimizations for the app to work properly.
        </div>
        <div class="content">
            <div class="wrapper">
                <div class="container">
                    <router-view/>
                </div>
            </div>
        </div>
        <div class="now-playing">
            <now-playing></now-playing>
        </div>
        <div class="controls">
            <controls :playbackData="playbackData"></controls>
        </div>
        <player @playback-data="onPlaybackData($event)"></player>
        <notifications class="notifications"></notifications>
        <media-session :playbackData="playbackData"></media-session>
    </div>
</template>

<style lang="scss" src="./App.scss"></style>

<script lang="ts" src="./App.ts"></script>
