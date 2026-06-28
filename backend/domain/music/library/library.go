package library

import (
	"sort"

	"github.com/boreq/eggplant/domain/music"
	"github.com/boreq/errors"
)

var (
	ErrAlbumNotFound     = errors.New("album not found")
	ErrTrackNotFound     = errors.New("track not found")
	ErrThumbnailNotFound = errors.New("thumbnail not found")
)

type Library struct {
	root RootAlbum
}

func NewLibrary(root RootAlbum) *Library {
	return &Library{root: root}
}

func (l *Library) GetRootAlbum(accessCtx AccessContext) (music.RootAlbum, error) {
	if !accessCtx.CanSee(l.getRootVisibility()) {
		return music.RootAlbum{}, errors.Wrap(ErrAlbumNotFound, "can't see root")
	}

	children := buildChildren(l.root.albums, l.getRootVisibilityToPropagateToChildren(), accessCtx)
	tracks := append([]music.Track(nil), l.root.tracks...)
	return music.NewRootAlbum(l.root.thumbnail, children, tracks)
}

func (l *Library) GetAlbum(accessCtx AccessContext, id music.AlbumId) (music.Album, error) {
	path, ok := l.findAlbumPath(id)
	if !ok {
		return music.Album{}, errors.Wrapf(ErrAlbumNotFound, "album '%s'", id)
	}

	effective := l.getRootVisibilityToPropagateToChildren()
	for _, a := range path {
		if a.visibility != nil {
			effective = *a.visibility
		}
	}

	if !accessCtx.CanSee(effective) {
		return music.Album{}, errors.Wrapf(ErrAlbumNotFound, "can't see album '%s'", id)
	}

	target := path[len(path)-1]

	ancestors := path[:len(path)-1]
	parents := make([]music.PartialAlbum, 0, len(ancestors))
	for _, a := range ancestors {
		parents = append(parents, music.NewPartialAlbum(a.id, a.title, a.thumbnail))
	}

	children := buildChildren(target.albums, effective, accessCtx)
	tracks := append([]music.Track(nil), target.tracks...)

	if len(children) == 0 && len(tracks) == 0 {
		return music.Album{}, errors.Wrapf(ErrAlbumNotFound, "album '%s' has no visible content", id)
	}

	return music.NewAlbum(target.id, target.title, target.thumbnail, parents, children, tracks)
}

func (l *Library) GetTrack(accessCtx AccessContext, id music.TrackId) (music.Track, error) {
	for _, t := range l.root.tracks {
		if t.Id() == id {
			if !accessCtx.CanSee(l.getRootVisibility()) {
				return music.Track{}, errors.Wrapf(ErrTrackNotFound, "can't see track '%s'", id)
			}
			return t, nil
		}
	}

	t, vis, ok := findTrackInAlbums(l.root.albums, id, l.getRootVisibilityToPropagateToChildren())
	if !ok {
		return music.Track{}, errors.Wrapf(ErrTrackNotFound, "track '%s' does not exist", id)
	}
	if !accessCtx.CanSee(vis) {
		return music.Track{}, errors.Wrapf(ErrTrackNotFound, "can't see track '%s'", id)
	}
	return t, nil
}

func (l *Library) GetThumbnail(accessCtx AccessContext, id music.ThumbnailId) (music.Thumbnail, error) {
	if t := l.root.thumbnail; t != nil && t.Id() == id {
		if !accessCtx.CanSee(l.getRootVisibility()) {
			return music.Thumbnail{}, errors.Wrapf(ErrThumbnailNotFound, "thumbnail '%s' not visible", id)
		}
		return *t, nil
	}

	t, vis, ok := findThumbnailInAlbums(l.root.albums, id, l.getRootVisibilityToPropagateToChildren())
	if !ok {
		return music.Thumbnail{}, errors.Wrapf(ErrThumbnailNotFound, "thumbnail '%s' does not exist", id)
	}
	if !accessCtx.CanSee(vis) {
		return music.Thumbnail{}, errors.Wrapf(ErrThumbnailNotFound, "can't see thumbnail '%s'", id)
	}
	return t, nil
}

func (l *Library) getRootVisibility() Visibility {
	return ifNotSetPublic(l.root.visibility)
}

func (l *Library) getRootVisibilityToPropagateToChildren() Visibility {
	return ifNotSetPrivate(l.root.visibility)
}

func findTrackInAlbums(albums []Album, id music.TrackId, parentVis Visibility) (music.Track, Visibility, bool) {
	for _, a := range albums {
		effective := parentVis
		if a.visibility != nil {
			effective = *a.visibility
		}
		for _, t := range a.tracks {
			if t.Id() == id {
				return t, effective, true
			}
		}
		if t, vis, ok := findTrackInAlbums(a.albums, id, effective); ok {
			return t, vis, true
		}
	}
	return music.Track{}, Visibility{}, false
}

func findThumbnailInAlbums(albums []Album, id music.ThumbnailId, parentVis Visibility) (music.Thumbnail, Visibility, bool) {
	for _, a := range albums {
		effective := parentVis
		if a.visibility != nil {
			effective = *a.visibility
		}
		if t := a.thumbnail; t != nil && t.Id() == id {
			return *t, effective, true
		}
		if t, vis, ok := findThumbnailInAlbums(a.albums, id, effective); ok {
			return t, vis, true
		}
	}
	return music.Thumbnail{}, Visibility{}, false
}

type RootAlbum struct {
	thumbnail  *music.Thumbnail
	visibility *Visibility
	albums     []Album
	tracks     []music.Track
}

func NewRootAlbum(thumbnail *music.Thumbnail, visibility *Visibility, albums []Album, tracks []music.Track) (RootAlbum, error) {
	return RootAlbum{
		thumbnail:  thumbnail,
		visibility: visibility,
		albums:     albums,
		tracks:     tracks,
	}, nil
}

func (r RootAlbum) Thumbnail() *music.Thumbnail {
	return r.thumbnail
}

func (r RootAlbum) Visibility() *Visibility {
	return r.visibility
}

func (r RootAlbum) Albums() []Album {
	return r.albums
}

func (r RootAlbum) Tracks() []music.Track {
	return r.tracks
}

type Album struct {
	id         music.AlbumId
	title      music.AlbumTitle
	thumbnail  *music.Thumbnail
	visibility *Visibility
	albums     []Album
	tracks     []music.Track
}

func NewAlbum(id music.AlbumId, title music.AlbumTitle, thumbnail *music.Thumbnail, visibility *Visibility, albums []Album, tracks []music.Track) (Album, error) {
	if len(albums) == 0 && len(tracks) == 0 {
		return Album{}, errors.New("album must have at least one child album or track")
	}
	return Album{
		id:         id,
		title:      title,
		thumbnail:  thumbnail,
		visibility: visibility,
		albums:     albums,
		tracks:     tracks,
	}, nil
}

func (a Album) Id() music.AlbumId {
	return a.id
}

func (a Album) Title() music.AlbumTitle {
	return a.title
}

func (a Album) Thumbnail() *music.Thumbnail {
	return a.thumbnail
}

func (a Album) Visibility() *Visibility {
	return a.visibility
}

func (a Album) Albums() []Album {
	return a.albums
}

func (a Album) Tracks() []music.Track {
	return a.tracks
}

func ifNotSetPrivate(v *Visibility) Visibility {
	if v == nil {
		return VisibilityPrivate
	}
	return *v
}

func ifNotSetPublic(v *Visibility) Visibility {
	if v == nil {
		return VisibilityPublic
	}
	return *v
}

func buildChildren(albums []Album, parentVis Visibility, ctx AccessContext) []music.PartialAlbum {
	var out []music.PartialAlbum
	for _, a := range albums {
		vis := parentVis
		if a.visibility != nil {
			vis = *a.visibility
		}
		if !ctx.CanSee(vis) {
			continue
		}
		out = append(out, music.NewPartialAlbum(a.id, a.title, a.thumbnail))
	}
	sortChildren(out)
	return out
}

func (l *Library) findAlbumPath(id music.AlbumId) ([]Album, bool) {
	return findAlbumPathRec(l.root.albums, id, nil)
}

func findAlbumPathRec(albums []Album, id music.AlbumId, path []Album) ([]Album, bool) {
	for _, a := range albums {
		path = append(path, a)
		if a.id == id {
			result := make([]Album, len(path))
			copy(result, path)
			return result, true
		}
		if result, ok := findAlbumPathRec(a.albums, id, path); ok {
			return result, true
		}
		path = path[:len(path)-1]
	}
	return nil, false
}

func sortChildren(albums []music.PartialAlbum) {
	sort.Slice(albums, func(i, j int) bool {
		return albums[i].Title().String() < albums[j].Title().String()
	})
}
