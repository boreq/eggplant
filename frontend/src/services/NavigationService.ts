import { Location } from 'vue-router';
import { BasicAlbum } from '@/dto/BasicAlbum';

export class NavigationService {

    getBrowse(album: BasicAlbum): Location {
        const ids = album.path ? album.path : [];
        const path = ids.join('/');
        return { path: `/browse/${path}` };
    }

}
