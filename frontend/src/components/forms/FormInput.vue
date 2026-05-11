<template>
    <div class="form-input">
        <div v-if="icon" class="icon">
            <i :class="icon"></i>
        </div>
        <input :type="type" :placeholder="placeholder"
               :value="value" @input="$emit('input', $event.target.value)" @keydown="onKeyDown">
    </div>
</template>
<script lang="ts" src="./FormInput.ts"></script>
<style scoped lang="scss" src="./FormInput.scss"></style>
