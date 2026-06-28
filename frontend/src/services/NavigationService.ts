import { Location } from 'vue-router';

export class NavigationService {

    getBrowse(albumId?: string, instanceId?: string): Location {
        if (albumId && instanceId) {
            return { name: 'remote-browse-children', params: { instanceId, albumId } };
        } else if (albumId) {
            return { name: 'browse-children', params: { albumId } };
        } else {
            return { name: 'browse' };
        }
    }

}
