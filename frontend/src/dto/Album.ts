import { Track } from '@/dto/Track';
import { Thumbnail } from '@/dto/Thumbnail';

export class Album {
    id?: string;
    title?: string;
    thumbnail?: Thumbnail;
    parents?: PartialAlbum[];
    albums: PartialAlbum[];
    tracks: Track[];
    remoteLibraryId?: string;
}

export class PartialAlbum {
    id: string;
    title: string;
    thumbnail?: Thumbnail;
    remoteLibraryId?: string;
}
