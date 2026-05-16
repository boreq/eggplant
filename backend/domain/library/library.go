package library

import (
	"sort"
	"strings"

	"github.com/boreq/eggplant/domain"
	"github.com/boreq/errors"
)

const maxSearchItems = 10

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

func (l *Library) GetRootAlbum(accessCtx AccessContext) (domain.RootAlbum, error) {
	if !accessCtx.CanSee(l.getRootVisibility()) {
		return domain.RootAlbum{}, errors.Wrap(ErrAlbumNotFound, "can't see root")
	}

	children := buildChildren(l.root.albums, l.getRootVisibilityToPropagateToChildren(), accessCtx)
	tracks := append([]domain.Track(nil), l.root.tracks...)
	return domain.NewRootAlbum(l.root.thumbnail, children, tracks)
}

func (l *Library) GetAlbum(accessCtx AccessContext, id domain.AlbumId) (domain.Album, error) {
	path, ok := l.findAlbumPath(id)
	if !ok {
		return domain.Album{}, errors.Wrapf(ErrAlbumNotFound, "album '%s'", id)
	}

	effective := l.getRootVisibilityToPropagateToChildren()
	for _, a := range path {
		if a.visibility != nil {
			effective = *a.visibility
		}
	}

	if !accessCtx.CanSee(effective) {
		return domain.Album{}, errors.Wrapf(ErrAlbumNotFound, "can't see album '%s'", id)
	}

	target := path[len(path)-1]

	parents := make([]domain.ParentAlbum, 0, len(path))
	for _, a := range path {
		parents = append(parents, domain.NewParentAlbum(a.id, a.title))
	}

	children := buildChildren(target.albums, effective, accessCtx)
	tracks := append([]domain.Track(nil), target.tracks...)

	if len(children) == 0 && len(tracks) == 0 {
		return domain.Album{}, errors.Wrapf(ErrAlbumNotFound, "album '%s' has no visible content", id)
	}

	return domain.NewAlbum(target.id, target.title, target.thumbnail, parents, children, tracks)
}

func (l *Library) GetTrack(accessCtx AccessContext, id domain.TrackId) (domain.Track, error) {
	for _, t := range l.root.tracks {
		if t.Id() == id {
			if !accessCtx.CanSee(l.getRootVisibility()) {
				return domain.Track{}, errors.Wrapf(ErrTrackNotFound, "can't see track '%s'", id)
			}
			return t, nil
		}
	}

	t, vis, ok := findTrackInAlbums(l.root.albums, id, l.getRootVisibilityToPropagateToChildren())
	if !ok {
		return domain.Track{}, errors.Wrapf(ErrTrackNotFound, "track '%s' does not exist", id)
	}
	if !accessCtx.CanSee(vis) {
		return domain.Track{}, errors.Wrapf(ErrTrackNotFound, "can't see track '%s'", id)
	}
	return t, nil
}

func (l *Library) GetThumbnail(accessCtx AccessContext, id domain.ThumbnailId) (domain.Thumbnail, error) {
	if t := l.root.thumbnail; t != nil && t.Id() == id {
		if !accessCtx.CanSee(l.getRootVisibility()) {
			return domain.Thumbnail{}, errors.Wrapf(ErrThumbnailNotFound, "thumbnail '%s' not visible", id)
		}
		return *t, nil
	}

	t, vis, ok := findThumbnailInAlbums(l.root.albums, id, l.getRootVisibilityToPropagateToChildren())
	if !ok {
		return domain.Thumbnail{}, errors.Wrapf(ErrThumbnailNotFound, "thumbnail '%s' does not exist", id)
	}
	if !accessCtx.CanSee(vis) {
		return domain.Thumbnail{}, errors.Wrapf(ErrThumbnailNotFound, "can't see thumbnail '%s'", id)
	}
	return t, nil
}

func (l *Library) Search(accessCtx AccessContext, query string) (SearchResult, error) {
	var result SearchResult

	if accessCtx.CanSee(l.getRootVisibility()) {
		for _, t := range l.root.tracks {
			if len(result.Tracks) >= maxSearchItems {
				break
			}
			if !containsStringCaseInsensitive(t.Title().String(), query) {
				continue
			}
			result.Tracks = append(result.Tracks, SearchTrack{Track: t})
		}
	}

	searchAlbums(&result, l.root.albums, nil, l.getRootVisibilityToPropagateToChildren(), query, accessCtx)
	return result, nil
}

func (l *Library) getRootVisibility() Visibility {
	return ifNotSetPublic(l.root.visibility)
}

func (l *Library) getRootVisibilityToPropagateToChildren() Visibility {
	return ifNotSetPrivate(l.root.visibility)
}

func findTrackInAlbums(albums []Album, id domain.TrackId, parentVis Visibility) (domain.Track, Visibility, bool) {
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
	return domain.Track{}, Visibility{}, false
}

func findThumbnailInAlbums(albums []Album, id domain.ThumbnailId, parentVis Visibility) (domain.Thumbnail, Visibility, bool) {
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
	return domain.Thumbnail{}, Visibility{}, false
}

type SearchAlbum struct {
	Path      []domain.AlbumId
	Title     domain.AlbumTitle
	Thumbnail *domain.Thumbnail
}

type SearchResult struct {
	Albums []SearchAlbum
	Tracks []SearchTrack
}

type SearchTrackAlbum struct {
	Path  []domain.AlbumId
	Title domain.AlbumTitle
}

type SearchTrack struct {
	Track domain.Track
	Album *SearchTrackAlbum
}

func searchAlbums(result *SearchResult, albums []Album, parentPath []domain.AlbumId, parentVis Visibility, query string, ctx AccessContext) {
	for _, a := range albums {
		effective := parentVis
		if a.visibility != nil {
			effective = *a.visibility
		}

		path := make([]domain.AlbumId, 0, len(parentPath)+1)
		path = append(path, parentPath...)
		path = append(path, a.id)

		if ctx.CanSee(effective) {
			if len(result.Albums) < maxSearchItems && containsStringCaseInsensitive(a.title.String(), query) {
				result.Albums = append(result.Albums, SearchAlbum{Path: path, Title: a.title, Thumbnail: a.thumbnail})
			}

			albumRef := &SearchTrackAlbum{Path: path, Title: a.title}
			for _, t := range a.tracks {
				if len(result.Tracks) >= maxSearchItems {
					break
				}
				if !containsStringCaseInsensitive(t.Title().String(), query) {
					continue
				}
				result.Tracks = append(result.Tracks, SearchTrack{Track: t, Album: albumRef})
			}
		}

		searchAlbums(result, a.albums, path, effective, query, ctx)
	}
}

type RootAlbum struct {
	thumbnail  *domain.Thumbnail
	visibility *Visibility
	albums     []Album
	tracks     []domain.Track
}

func NewRootAlbum(thumbnail *domain.Thumbnail, visibility *Visibility, albums []Album, tracks []domain.Track) (RootAlbum, error) {
	return RootAlbum{
		thumbnail:  thumbnail,
		visibility: visibility,
		albums:     albums,
		tracks:     tracks,
	}, nil
}

func (r RootAlbum) Thumbnail() *domain.Thumbnail {
	return r.thumbnail
}

func (r RootAlbum) Visibility() *Visibility {
	return r.visibility
}

func (r RootAlbum) Albums() []Album {
	return r.albums
}

func (r RootAlbum) Tracks() []domain.Track {
	return r.tracks
}

type Album struct {
	id         domain.AlbumId
	title      domain.AlbumTitle
	thumbnail  *domain.Thumbnail
	visibility *Visibility
	albums     []Album
	tracks     []domain.Track
}

func NewAlbum(id domain.AlbumId, title domain.AlbumTitle, thumbnail *domain.Thumbnail, visibility *Visibility, albums []Album, tracks []domain.Track) (Album, error) {
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

func (a Album) Id() domain.AlbumId {
	return a.id
}

func (a Album) Title() domain.AlbumTitle {
	return a.title
}

func (a Album) Thumbnail() *domain.Thumbnail {
	return a.thumbnail
}

func (a Album) Visibility() *Visibility {
	return a.visibility
}

func (a Album) Albums() []Album {
	return a.albums
}

func (a Album) Tracks() []domain.Track {
	return a.tracks
}

func ifNotSetPrivate(v *Visibility) Visibility {
	if v == nil {
		return defaultVisibility
	}
	return *v
}

func ifNotSetPublic(v *Visibility) Visibility {
	if v == nil {
		return NewVisibility(true)
	}
	return *v
}

func buildChildren(albums []Album, parentVis Visibility, ctx AccessContext) []domain.ChildAlbum {
	var out []domain.ChildAlbum
	for _, a := range albums {
		vis := parentVis
		if a.visibility != nil {
			vis = *a.visibility
		}
		if !ctx.CanSee(vis) {
			continue
		}
		out = append(out, domain.NewChildAlbum(a.id, a.title, a.thumbnail))
	}
	sortChildren(out)
	return out
}

func (l *Library) findAlbumPath(id domain.AlbumId) ([]Album, bool) {
	return findAlbumPathRec(l.root.albums, id, nil)
}

func findAlbumPathRec(albums []Album, id domain.AlbumId, path []Album) ([]Album, bool) {
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

func sortChildren(albums []domain.ChildAlbum) {
	sort.Slice(albums, func(i, j int) bool {
		return albums[i].Title().String() < albums[j].Title().String()
	})
}

func containsStringCaseInsensitive(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
