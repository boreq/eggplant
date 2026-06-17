<template>
    <div class="subheader" :class="{ clickable }" @click="$emit('click')">
        <div class="text">{{ text }}</div>
        <div class="line"></div>
        <div class="action" v-if="$slots.action">
            <slot name="action"></slot>
        </div>
    </div>
</template>
<script lang="ts" src="./SubHeader.ts"></script>
<style scoped lang="scss" src="./SubHeader.scss"></style>
