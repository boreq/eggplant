<template>
    <div class="controls">
        <div class="core">
            <div class="shuffle">
                <indicator :state="shuffle">
                    <a @click="onShuffle()">
                        <i class="fas fa-random"></i>
                    </a>
                </indicator>
            </div>

            <div class="previous">
                <a @click="onPrevious()">
                    <i class="fas fa-backward"></i>
                </a>
            </div>

            <div class="play">
                <a @click="onPlayPause()">
                    <indicator :state="!paused" :width="15">
                        Play
                    </indicator>
                </a>
            </div>

            <div class="next">
                <a @click="onNext()">
                    <i class="fas fa-forward"></i>
                </a>
            </div>

            <div class="volume">
                <Volume></Volume>
            </div>
        </div>

        <div class="progress">
            <div class="current">
                {{ currentTime }}
            </div>

            <div class="bar">
                <progress-bar :value="currentTimePercentage" @value-selected="seek"></progress-bar>
            </div>

            <div class="duration">
                {{ duration }}
            </div>
        </div>
    </div>
</template>
<script lang="ts" src="./Controls.ts"></script>
<style scoped lang="scss" src="./Controls.scss"></style>
