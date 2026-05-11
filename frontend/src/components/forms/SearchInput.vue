<template>
    <div class="search-input">
        <div class="icon">
            <i class="fas fa-search"></i>
        </div>
            <a class="close" v-if="showCloseButton" @click="$emit('input', '')">
                <i class="fas fa-times"></i>
            </a>
        <input type="text" placeholder="Search"
               maxlength="100" :value="value"
               @input="$emit('input', $event.target.value)"
               @keydown.stop>
    </div>
</template>
<script lang="ts" src="./SearchInput.ts"></script>
<style scoped lang="scss" src="./SearchInput.scss"></style>
