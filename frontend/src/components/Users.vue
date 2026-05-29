<template>
    <div class="users">
        <table>
            <thead>
            <th class="username">
                Username
            </th>
            <th class="role">
                Role
            </th>
            <th class="created">
                Signed Up
            </th>
            <th class="last-seen">
                Last Seen
            </th>
            <th class="actions">
                Actions
            </th>
            </thead>
            <tbody>
            <tr class="user" v-for="user in users" :key="user.username">
                <td class="username">
                    {{ user.username }}
                </td>

                <td class="role">
                    {{ getRole(user) }}
                </td>

                <td class="created">
                    <timeago v-if="user.created" :datetime="user.created"></timeago>
                    <template v-else>Unknown</template>
                </td>

                <td class="last-seen">
                    <timeago v-if="user.lastSeen" :datetime="user.lastSeen"></timeago>
                    <template v-else>Never</template>
                </td>

                <td class="actions">
                    <v-popover offset="16">
                        <a v-tooltip="'Remove this user.'" v-if="canRemove(user)">
                            <i class="fas fa-times"></i>
                        </a>

                        <template slot="popover">
                            <p>
                                This action can not be undone, do you wish to proceed?
                            </p>
                            <app-button text="Confirm" v-close-popover @click="removeUser(user)"></app-button>
                        </template>
                    </v-popover>
                </td>
            </tr>
            </tbody>
        </table>
        <spinner v-if="!users"></spinner>
    </div>
</template>
<script lang="ts" src="./Users.ts"></script>
<style scoped lang="scss" src="./Users.scss"></style>
