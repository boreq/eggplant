<template>
    <ul class="tracks">
        <li v-for="(entry, index) in entries" :key="entry.track.id" class="track" :class="{ playing: isNowPlaying(index, entry.track) }" @click="onTrackRowClicked(index)">
            <div class="play">
                <a @click="onPlayTrackButtonClicked(index)" class="button" v-if="!showPlayAlbumButtonAsPause(index, entry.track)">
                    <i class="icon fas fa-play"></i>
                </a>

                <a @click="onPlayTrackButtonClicked(index)" class="button" v-else>
                    <i class="icon fas fa-pause"></i>
                </a>

                <div class="number">
                    <template v-if="!hideNumber && entry.track.number != null">{{ entry.track.number }}</template>
                    <i v-else class="icon fas fa-music"></i>
                </div>
            </div>

            <div class="title" :title="entry.track.title">
                {{ entry.track.title }}
            </div>

            <dropdown class="actions" ref="dropdowns" @click.native.stop="">
                <template v-slot:trigger>
                    <div class="target">
                        <i class="fas fa-ellipsis-h"></i>
                    </div>
                </template>
                <template v-slot:content>
                    <dropdown-element>
                        <a @click="addToQueue(entry)">
                            Add to queue
                        </a>
                    </dropdown-element>
                    <dropdown-element v-if="queueMode">
                        <a @click="removeFromQueue(entry)">
                            Remove from queue
                        </a>
                    </dropdown-element>
                    <dropdown-divider v-if="showGoToAlbum && entry.album"></dropdown-divider>
                    <dropdown-element v-if="showGoToAlbum && entry.album">
                        <a @click="goToAlbum(entry)">
                            Go to album
                        </a>
                    </dropdown-element>
                </template>
            </dropdown>

            <div class="duration">
                {{ formatDuration(entry.track) }}
            </div>
        </li>
    </ul>
</template>
<script lang="ts" src="./Tracks.ts"></script>
<style scoped lang="scss" src="./Tracks.scss"></style>
