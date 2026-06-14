export interface Track {
    id: string;
    number?: number;
    title: string;
    // Loaded lazily from the track duration endpoint; undefined until fetched.
    duration?: number;
    // Set when this track comes from a linked remote instance.
    remoteId?: string;
}
