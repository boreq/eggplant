<template>
    <div class="vertical-progress-bar" ref="progress-bar"
         @click="onClick" @mousemove="onMouseMove" @mousedown="onMouseDown" @mouseup="onMouseUp" @mouseleave="onMouseLeave">
        <div class="bar">
            <div class="filled-in" :style="{ height: percentage }">
            </div>
        </div>
    </div>
</template>
<script lang="ts" src="./VerticalProgressBar.ts"></script>
<style scoped lang="scss" src="./VerticalProgressBar.scss"></style>
