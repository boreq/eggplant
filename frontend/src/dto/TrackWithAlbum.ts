import { Track } from '@/dto/Track';
import { PartialAlbum } from "@/dto/Album";

export class TrackWithAlbum {
    album?: PartialAlbum;
    track: Track;
}
