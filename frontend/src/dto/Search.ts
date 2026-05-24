import { Track } from '@/dto/Track';
import {PartialAlbum} from "@/dto/Album";

export class SearchResults {
    albums: PartialAlbum[];
    tracks: SearchResultTrack[];
}

export class SearchResultTrack {
    track: Track;
    album?: PartialAlbum;
}
