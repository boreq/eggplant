<template>
    <div class="dropdown">
        <div class="close-overlay" @click="close" v-if="isOpen"></div>
        <div class="black-overlay" :class="{ dark: isOpen }"></div>
        <div class="wrapper">
            <a class="trigger" @click="open" ref="trigger">
                <slot name="trigger"></slot>
            </a>
            <div class="arrow" v-if="isOpen" :style="{ top: arrowTop, left: arrowLeft }"></div>
            <ul class="content" v-if="isOpen" :style="{ top: top, left: left }">
                <slot name="content"></slot>
            </ul>
        </div>
    </div>
</template>
<script lang="ts" src="./Dropdown.ts"></script>
<style scoped lang="scss" src="./Dropdown.scss"></style>
