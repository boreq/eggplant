<template>
    <div class="browse" :class="{forbidden: forbidden}">
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
                <spinner v-if="!album" class="spinner"></spinner>

                <div class="album" v-if="showAlbum">
                    <div class="artwork">
                        <thumbnail :album="basicAlbum"></thumbnail>
                    </div>

                    <div class="info">
                        <ul class="crumbs">
                            <li>
                                <router-link to="/browse">
                                    Eggplant
                                </router-link>
                            </li>
                            <li v-for="parent in album.parents">
                                <router-link :to="parentUrl(parent)">
                                    {{ parent.title }}
                                </router-link>
                            </li>
                        </ul>

                        <div class="title">
                            {{ album.title }}
                        </div>

                        <div class="details" v-if="numberOfTracks > 0">
                            {{ numberOfTracks }} tracks, {{ totalDurationMinutes }} minutes
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

                <div v-if="album && album.tracks">
                    <SubHeader text="Tracks"></SubHeader>
                    <Tracks :entries="entries"></Tracks>
                </div>

                <div v-if="albums">
                    <SubHeader text="Albums"></SubHeader>
                    <Albums :albums="albums" @select-album="selectAlbum"></Albums>
                </div>

                <div v-if="album && !album.albums && !album.tracks" class="no-content-message">
                    The music library is empty or you do not have permissions
                    to view any tracks or albums.
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

            <div class="forbidden-message">
                <div class="message">
                    <div class="icon">
                        <i class="fas fa-exclamation-triangle"></i>
                    </div>
                    Access denied, please confirm that you are signed in.
                </div>
            </div>
        </div>
    </div>
</template>
<script lang="ts" src="./Browse.ts"></script>
<style scoped lang="scss" src="./Browse.scss"></style>
