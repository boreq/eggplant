<template>
    <div class="albums">
        <a v-for="album in albums" class="album" @click="selectAlbum(album)">
            <div class="image">
                <thumbnail :album="album"></thumbnail>
            </div>
            <div class="title">
                {{ album.title }}
            </div>
        </a>
    </div>
</template>
<script lang="ts" src="./Albums.ts"></script>
<style scoped lang="scss" src="./Albums.scss"></style>
