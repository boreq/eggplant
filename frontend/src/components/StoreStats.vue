<template>
    <div class="store-stats">
        <div>
            <div class="label">
               Number of items
            </div>
            <div class="value">
                {{ stats.allItems }}
            </div>
        </div>
        <div>
            <div class="label">
                Conversion progress
            </div>
            <div class="value">
                {{ conversionProgress }}
            </div>
        </div>
        <div>
            <div class="label">
                Original size
            </div>
            <div class="value">
                {{ originalSize }}
            </div>
        </div>
        <div>
            <div class="label">
                Converted size
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
<script lang="ts" src="./StoreStats.ts"></script>
<style scoped lang="scss" src="./StoreStats.scss"></style>
