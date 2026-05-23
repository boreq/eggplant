<template>
    <div class="store-stats">
        <div>
            <div class="label">
                Number of thumbnails
            </div>
            <div class="value">
                {{ stats.allItems }}
            </div>
        </div>
        <div>
            <div class="label">
                Number of converted thumbnails
            </div>
            <div class="value">
                {{ stats.convertedItems }}
            </div>
        </div>
        <div>
            <div class="label">
                Size of thumbnails
            </div>
            <div class="value">
                {{ originalSize }}
            </div>
        </div>
        <div>
            <div class="label">
                Size of converted thumbnails
            </div>
            <div class="value">
                {{ convertedSize }}
            </div>
        </div>
        <div>
            <div class="label">
                Conversion ratio
            </div>
            <div class="value">
                {{ conversionRatio }}
            </div>
        </div>
    </div>
</template>
<script lang="ts" src="./ThumbnailStats.ts"></script>
<style scoped lang="scss" src="./StoreStats.scss"></style>
