<template>
    <div class="thumbnail" :class="{ tilt: tilt }" @mousemove="onMouseMove" @mouseleave="onMouseLeave" ref="root">
        <div class="artwork" v-if="showImage">
            <img :src="thumbnailUrl" :class="{ loaded }" @error="onError" @load="onLoad">
            <div class="glare" v-if="tilt" aria-hidden="true"></div>
            <div class="spinner-container" v-if="!loaded">
                <spinner></spinner>
            </div>
        </div>
        <div class="icon" v-else>
            <i class="fas fa-compact-disc"></i>
        </div>
    </div>
</template>
<script lang="ts" src="./Thumbnail.ts"></script>
<style scoped lang="scss" src="./Thumbnail.scss"></style>
