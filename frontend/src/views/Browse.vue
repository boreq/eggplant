<template>
    <div class="browse" :class="state">
        <div class="wrapper">
            <div class="topbar">
                <search-input class="search" v-model="searchQuery"></search-input>

                <ul class="buttons">
                    <li class="button" :class="{ active: showQueue }">
                        <a @click="toggleQueue" v-tooltip="showQueue ? 'Close queue.' : 'Queue.'">
                            <i class="fas fa-list-ol"></i>
                        </a>
                    </li>
                    <li class="button">
                        <login-button></login-button>
                    </li>
                </ul>
            </div>

            <div class="content" ref="content">
                <div class="album-hero" v-if="showAlbum">
                    <div class="artwork">
                        <thumbnail :album="basicAlbum" tilt></thumbnail>
                    </div>

                    <div class="info">
                        <ul class="crumbs">
                            <li>
                                <router-link :to="{ name: 'browse' }">
                                    Eggplant
                                </router-link>
                            </li>
                            <li v-if="remoteInstanceId">
                                <router-link v-if="album.id" :to="remoteInstanceUrl">
                                    {{ remoteInstanceTitle }}
                                </router-link>
                                <template v-else>
                                    {{ remoteInstanceTitle }}
                                </template>
                            </li>
                            <li v-for="parent in album.parents" :key="parent.id">
                                <router-link :to="parentUrl(parent)">
                                    {{ parent.title }}
                                </router-link>
                            </li>
                            <li v-if="album.id">
                                {{ album.title }}
                            </li>
                        </ul>

                        <div class="title">
                            {{ title }}
                        </div>

                        <div class="details" v-if="numberOfTracks > 0">
                            {{ numberOfTracks }} tracks, {{ totalDurationMinutes }}<template v-if="durationsLoading">+</template> minutes
                        </div>

                        <ul class="actions" v-if="numberOfTracks > 0">
                            <li>
                                <a class="play" @click="onPlayAlbumButtonClicked">
                                    <i class="icon fas fa-play" v-if="!showPlayAlbumButtonAsPause"></i>
                                    <i class="icon fas fa-pause" v-else></i>
                                </a>
                            </li>
                            <li>
                                <dropdown ref="dropdown">
                                    <template v-slot:trigger>
                                        <a class="secondary">
                                            <i class="fas fa-ellipsis-h"></i>
                                        </a>
                                    </template>
                                    <template v-slot:content>
                                        <dropdown-element>
                                            <a @click="addAlbumToQueue">
                                                Add album to queue
                                            </a>
                                        </dropdown-element>
                                    </template>
                                </dropdown>
                            </li>
                        </ul>
                    </div>
                </div>

                <div class="section" v-if="showRemoteLibraries">
                    <SubHeader text="Connected instances"></SubHeader>
                    <remote-libraries :libraries="remoteLibraries"
                        @select-library="selectRemoteLibrary"></remote-libraries>
                </div>

                <div class="section bleed" v-if="album && album.tracks && album.tracks.length > 0">
                    <SubHeader text="Tracks" :clickable="canToggleTracks"
                        @click="canToggleTracks && (showAllTracks = !showAllTracks)">
                        <template v-slot:action v-if="canToggleTracks">
                            <a>
                                {{ showAllTracks ? 'Show fewer' : 'View all ' + entries.length }}
                            </a>
                        </template>
                    </SubHeader>
                    <Tracks :entries="displayedEntries" :allEntries="entries" :highlightTrackId="highlightTrackId"></Tracks>
                </div>

                <div class="section" v-if="albums && albums.length > 0">
                    <SubHeader text="Albums" :clickable="canToggleAlbums"
                        @click="canToggleAlbums && (showAllAlbums = !showAllAlbums)">
                        <template v-slot:action v-if="canToggleAlbums">
                            <a>
                                {{ showAllAlbums ? 'Show fewer' : 'View all ' + albums.length }}
                            </a>
                        </template>
                    </SubHeader>
                    <Albums :albums="displayedAlbums" @select-album="selectAlbum"></Albums>
                </div>
            </div>

            <div class="content queue" v-if="showQueue">
                <main-header text="Queue"></main-header>
                <queue @select-album="selectAlbum"></queue>
            </div>

            <div class="content search" v-if="showSearch">
                <main-header text="Search"></main-header>
                <search :query="searchQuery" @select-album="selectAlbum"></search>
            </div>

            <div class="message-overlay empty-message">
                <div class="message">
                    <div class="icon">
                        <i class="fas fa-music"></i>
                    </div>
                    The music library is empty or you do not have permission to view any tracks or albums.
                </div>
            </div>

            <div class="message-overlay loading-message">
                <div class="message">
                    <div class="icon">
                        <spinner></spinner>
                    </div>
                </div>
            </div>

            <div class="message-overlay permission-denied-message">
                <div class="message">
                    <div class="icon">
                        <i class="fas fa-exclamation-triangle"></i>
                    </div>
                    The music library may be empty, this album may not exist or you may not have permission to access
                    any tracks or albums.
                </div>
            </div>

            <div class="message-overlay library-not-ready-message">
                <div class="message">
                    <div class="icon">
                        <spinner></spinner>
                    </div>
                    The music library is being prepared.
                    This page will load automatically when it's ready.
                </div>
            </div>
        </div>
    </div>
</template>
<script lang="ts" src="./Browse.ts"></script>
<style scoped lang="scss" src="./Browse.scss"></style>
