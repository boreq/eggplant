import { Track } from '@/dto/Track';
import {PartialAlbum} from "@/dto/Album";

export class SearchResults {
    albums: SearchResultAlbum[];
    tracks: SearchResultTrack[];
}

export class SearchResultAlbum {
    album: PartialAlbum;
    score: number;
}

export class SearchResultTrack {
    track: Track;
    album?: PartialAlbum;
    score: number;
}
