<template>
    <div class="version">
        <spinner v-if="version === null"></spinner>
        <p v-else-if="frontendVersion === backendVersion">This is Eggplant version <code>{{ frontendVersion }}</code>.</p>
        <p v-else>This is frontend version <code>{{ frontendVersion }}</code> and backend version <code>{{ backendVersion }}</code>.</p>
    </div>
</template>
<script lang="ts" src="./Version.ts"></script>
<style scoped lang="scss" src="./Version.scss"></style>
