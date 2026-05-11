import { Track } from '@/dto/Track';
import { Thumbnail } from '@/dto/Thumbnail';
import { Access } from '@/dto/Access';

export class Album {
    id: string;
    title: string;
    thumbnail: Thumbnail;
    parents: Album[];
    albums: Album[];
    tracks: Track[];
    access: Access;
}
