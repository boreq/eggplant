<template>
    <div class="thumbnail">
        <div class="artwork" v-if="album.thumbnail">
            <img :src="thumbnailUrl" @error="onError" @load="onLoad" ref="image">
            <div class="spinner-container" v-if="converting" v-tooltip="'This thumbnail has not been converted yet.'">
                <spinner></spinner>
            </div>
        </div>
        <div class="icon" v-if="!album.thumbnail">
            <i class="fas fa-compact-disc"></i>
        </div>
    </div>
</template>
<script lang="ts" src="./Thumbnail.ts"></script>
<style scoped lang="scss" src="./Thumbnail.scss"></style>
