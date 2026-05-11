<template>
    <div class="stats">
        <action-bar>
            <action-bar-button text="Go back to settings" icon="fas fa-arrow-left" @click="goToSettings"></action-bar-button>
        </action-bar>
        <main-header text="Statistics"></main-header>
        <spinner v-if="!stats" class="spinner"></spinner>
        <div v-if="stats">
            <sub-header text="Tracks"></sub-header>
            <store-stats :stats="stats.tracks"></store-stats>

            <sub-header text="Thumbnails"></sub-header>
            <store-stats :stats="stats.thumbnails"></store-stats>
        </div>
    </div>
</template>
<script lang="ts" src="./Stats.ts"></script>
<style scoped lang="scss" src="./Stats.scss"></style>
