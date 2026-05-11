export class Stats {
    users: number;
    thumbnails: StoreStats;
    tracks: StoreStats;
}

export class StoreStats {
    allItems: number;
    convertedItems: number;
    originalSize: number;
    convertedSize: number;
}
