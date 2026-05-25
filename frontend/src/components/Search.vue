<template>
    <div class="search">
        <spinner v-if="!result"></spinner>

        <div class="results" v-if="result">
            <div class="section bleed">
                <SubHeader text="Tracks">
                    <template v-slot:action v-if="canToggleTracks">
                        <a @click="showAllTracks = !showAllTracks">
                            {{ showAllTracks ? 'Show fewer' : 'View all ' + tracks.length }}
                        </a>
                    </template>
                </SubHeader>
                <Tracks v-if="hasTracks" :entries="displayedTracks" :showGoToAlbum="true" :hideNumber="true" @select-album="selectAlbum"></Tracks>
                <div class="message" v-else>
                    No tracks found.
                </div>
            </div>

            <div class="section">
                <SubHeader text="Albums">
                    <template v-slot:action v-if="canToggleAlbums">
                        <a @click="showAllAlbums = !showAllAlbums">
                            {{ showAllAlbums ? 'Show fewer' : 'View all ' + albums.length }}
                        </a>
                    </template>
                </SubHeader>
                <Albums v-if="hasAlbums" :albums="displayedAlbums" @select-album="selectAlbum"></Albums>
                <div class="message" v-else>
                    No albums found.
                </div>
            </div>
        </div>
    </div>
</template>
<script lang="ts" src="./Search.ts"></script>
<style scoped lang="scss" src="./Search.scss"></style>
