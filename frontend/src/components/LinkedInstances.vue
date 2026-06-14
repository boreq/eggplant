<template>
    <div class="linked-instances">
        <div v-if="loading" class="loading">Loading...</div>

        <ul v-else-if="instances.length > 0" class="instance-list">
            <li v-for="inst in instances" :key="inst.id" class="instance-item">
                <span class="instance-name">{{ inst.name }}</span>
                <span class="instance-url">{{ inst.url }}</span>
                <button class="remove-btn" @click="remove(inst)">Remove</button>
            </li>
        </ul>
        <p v-else class="no-instances">No linked instances yet.</p>

        <div class="add-form">
            <h4>Link a new instance</h4>
            <form-input v-model="addName" placeholder="Name (e.g. Alice's library)"></form-input>
            <form-input v-model="addUrl" placeholder="URL (e.g. https://music.example.com)"></form-input>
            <form-input v-model="addUsername" placeholder="Your username on that instance"></form-input>
            <form-input v-model="addPassword" type="password" placeholder="Your password on that instance"></form-input>
            <app-button text="Link instance" @click="add" :working="addInProgress"></app-button>
        </div>
    </div>
</template>
<script lang="ts" src="./LinkedInstances.ts"></script>
<style scoped lang="scss" src="./LinkedInstances.scss"></style>
