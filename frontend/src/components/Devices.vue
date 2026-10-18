<template>
    <div class="devices">
        <sub-header text="This device"></sub-header>
        <div class="rename">
            <form-input class="input" v-model="name" placeholder="Device name" icon="fas fa-pen"
                        @submit="rename"></form-input>
            <app-button text="Rename" :disabled="!nameChanged" :working="renaming" @click="rename"></app-button>
        </div>

        <sub-header text="Devices"></sub-header>
        <div class="message" v-if="devices.length === 0">
            No devices are online.
        </div>
        <ul class="list">
            <li class="device" v-for="device in devices" :key="device.id"
                :class="{ selectable: canSelect(device), current: isCurrent(device) }"
                @click="select(device)">
                <div class="icon">
                    <i class="fas" :class="icon(device)"></i>
                </div>
                <div class="info">
                    <div class="name">{{ device.name }}</div>
                    <div class="description">{{ describe(device) }}</div>
                </div>
            </li>
        </ul>
    </div>
</template>
<script lang="ts" src="./Devices.ts"></script>
<style scoped lang="scss" src="./Devices.scss"></style>
