import { Location } from 'vue-router';

export class NavigationService {

    getBrowse(albumId?: string): Location {
        if (albumId) {
            return { name: `browse-children`, params: { id: albumId } };
        } else {
            return { name: `browse` };
        }
    }

}
