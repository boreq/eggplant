import { Location } from 'vue-router';

export class NavigationService {

    getBrowse(albumId?: string, libraryId?: string): Location {
        if (albumId && libraryId) {
            return { name: 'library-browse-children', params: { libraryId, albumId } };
        } else if (libraryId) {
            return { name: 'library-browse', params: { libraryId } };
        } else if (albumId) {
            return { name: 'browse-children', params: { albumId } };
        } else {
            return { name: 'browse' };
        }
    }

}
