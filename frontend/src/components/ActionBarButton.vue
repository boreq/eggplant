<template>
    <div class="action-bar-button">
        <a @click="onClick" v-if="!working">
            <i :class="icon"></i> {{ text }}
        </a>
        <spinner v-if="working"></spinner>
    </div>
</template>
<script lang="ts" src="./ActionBarButton.ts"></script>
<style scoped lang="scss" src="./ActionBarButton.scss"></style>
