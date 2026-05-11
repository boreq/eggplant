<template>
    <div class="now-playing" :class="{ hidden: !nowPlaying}">
        <div class="trackTitle" v-if="nowPlaying">
            <a @click="goToNowPlayingSong()">
                {{ trackTitle }}
            </a>
        </div>
        <div class="albumTitle" v-if="nowPlaying">
            <a @click="goToNowPlayingAlbum()">
                {{ albumTitle }}
            </a>
        </div>
    </div>
</template>
<script lang="ts" src="./NowPlaying.ts"></script>
<style scoped lang="scss" src="./NowPlaying.scss"></style>
