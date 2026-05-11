import { Track } from '@/dto/Track';
import { BasicAlbum } from '@/dto/BasicAlbum';

export class SearchResult {
    albums: BasicAlbum[];
    tracks: SearchResultTrack[];
}

export class SearchResultTrack {
    track: Track;
    album: BasicAlbum;
}
