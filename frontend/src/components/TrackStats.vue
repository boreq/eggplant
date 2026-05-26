<template>
    <div class="store-stats">
        <div>
            <div class="label">
                Number of tracks
            </div>
            <div class="value">
                {{ stats.numberOfTracks }}
            </div>
        </div>
        <div>
            <div class="label">
                Size of tracks
            </div>
            <div class="value">
                {{ sizeOfTracks }}
            </div>
        </div>
        <div>
            <div class="label">
                Number of streams
            </div>
            <div class="value">
                {{ stats.numberOfStreams }}
            </div>
        </div>
        <div>
            <div class="label">
                Size of stream cache
            </div>
            <div class="value">
                {{ sizeOfConvertedTracks }}
            </div>
        </div>
    </div>
</template>
<script lang="ts" src="./TrackStats.ts"></script>
<style scoped lang="scss" src="./StoreStats.scss"></style>
