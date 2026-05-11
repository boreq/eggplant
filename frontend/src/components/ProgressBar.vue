<template>
    <div class="progress-bar" ref="progress-bar"
         @click="onClick" @mousemove="onMouseMove" @mousedown="onMouseDown" @mouseup="onMouseUp" @mouseleave="onMouseLeave">
        <div class="bar">
            <div class="filled-in" :style="{ width: percentage }">
            </div>
        </div>
    </div>
</template>
<script lang="ts" src="./ProgressBar.ts"></script>
<style scoped lang="scss" src="./ProgressBar.scss"></style>
