<template>
    <ul class="notifications">
        <li class="notification" :class="[notification.class, shouldHide(notification) ? 'hide': '']" v-for="notification of notifications" :key="notification.id">
            <div class="text">
                {{ notification.text }}
            </div>
            <div class="extra" v-if="notification.extra">
                {{ notification.extra }}
            </div>
        </li>
    </ul>
</template>
<script lang="ts" src="./Notifications.ts"></script>
<style scoped lang="scss" src="./Notifications.scss"></style>
