export class Stats {
    users: number;
    thumbnails: ThumbnailStats;
    tracks: TrackStats;
}

export class TrackStats {
    numberOfTracks: number;
    sizeOfTracks: number;
    numberOfStreams: number;
    sizeOfConvertedTracks: number;
}

export class ThumbnailStats {
    allItems: number;
    convertedItems: number;
    originalSize: number;
    convertedSize: number;
}
