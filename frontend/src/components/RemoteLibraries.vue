<template>
    <div class="remote-libraries">
        <a v-for="library in libraries" :key="library.id" class="library" @click="selectLibrary(library)">
            <div class="icon">
                <i class="fas fa-network-wired"></i>
            </div>
            <div class="name">
                {{ library.name }}
            </div>
        </a>
    </div>
</template>
<script lang="ts" src="./RemoteLibraries.ts"></script>
<style scoped lang="scss" src="./RemoteLibraries.scss"></style>
