<template>
    <div class="search">
        <spinner v-if="!result"></spinner>

        <div v-if="result">
            <SubHeader text="Tracks"></SubHeader>
            <Tracks :entries="tracks" :showGoToAlbum="true" @select-album="selectAlbum"></Tracks>

            <div class="message" v-if="tracks.length === 0">
                No tracks found.
            </div>

            <SubHeader text="Albums"></SubHeader>
            <Albums :albums="albums" @select-album="selectAlbum"></Albums>

            <div class="message" v-if="albums.length === 0">
                No albums found.
            </div>
        </div>
    </div>
</template>
<script lang="ts" src="./Search.ts"></script>
<style scoped lang="scss" src="./Search.scss"></style>
