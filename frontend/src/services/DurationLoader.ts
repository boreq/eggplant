import Vue from 'vue';
import { Track } from '@/dto/Track';
import { ApiService } from '@/services/ApiService';

// Loads track durations lazily from the per-track duration endpoint, in
// batches, so a large album does not fire hundreds of requests at once. Owned
// by the component that owns the track list; presentation components only
// render whatever duration ends up set.
export class DurationLoader {

    private readonly batchSize = 10;
    private readonly apiService: ApiService;

    private generation = 0;
    private tracks: Track[] = [];

    constructor(vue: unknown) {
        this.apiService = new ApiService(vue);
    }

    load(tracks: Track[]): void {
        this.tracks = tracks || [];
        const generation = ++this.generation;
        this.loadNextBatch(generation);
    }

    cancel(): void {
        this.generation++;
    }

    private loadNextBatch(generation: number): void {
        if (generation !== this.generation) {
            return;
        }
        const batch = this.tracks
            .filter(track => track.duration === undefined)
            .slice(0, this.batchSize);
        if (batch.length === 0) {
            return;
        }
        const requests = batch.map(track => this.apiService.getTrackDuration(track)
            .then(response => Vue.set(track, 'duration', response.data.duration))
            .catch(() => { /* duration stays unknown; rendered blank */ }));
        Promise.all(requests).then(() => this.loadNextBatch(generation));
    }

}
