import { Track } from '@/dto/Track';
import { Thumbnail } from '@/dto/Thumbnail';

export class Album {
    id: string;
    title: string;
    thumbnail: Thumbnail;
    parents: Album[];
    albums: Album[];
    tracks: Track[];
}
