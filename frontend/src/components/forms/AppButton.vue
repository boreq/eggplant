<template>
    <a class="app-button" :class="{ disabled: disabled, working: working }" @click="onClick">
        <span v-if="!working">{{ text }}</span>
        <spinner v-if="working"></spinner>
    </a>
</template>
<script lang="ts" src="./AppButton.ts"></script>
<style scoped lang="scss" src="./AppButton.scss"></style>
