<template>
    <div class="indicator">
        <div class="bar" :class="{ enabled: state }" :style="{ width: barWidth }">
        </div>
        <div class="content">
            <slot></slot>
        </div>
    </div>
</template>
<script lang="ts" src="./Indicator.ts"></script>
<style scoped lang="scss" src="./Indicator.scss"></style>
