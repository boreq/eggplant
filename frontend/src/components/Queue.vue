<template>
    <div class="queue">
        <div class="message" v-if="empty">
            Your queue is empty.
        </div>

        <tracks :entries="entries" :queueMode="true" :showGoToAlbum="true" @select-album="onSelectAlbum"></tracks>
    </div>
</template>
<script lang="ts" src="./Queue.ts"></script>
<style scoped lang="scss" src="./Queue.scss"></style>
