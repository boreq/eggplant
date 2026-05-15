package library

import (
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/boreq/eggplant/domain"
	"github.com/boreq/errors"
)

const maxSearchItems = 10

var (
	ErrAlbumNotFound     = errors.New("album not found")
	ErrTrackNotFound     = errors.New("track not found")
	ErrThumbnailNotFound = errors.New("thumbnail not found")
)

var rootAlbumTitle = mustAlbumTitle("Eggplant")
var defaultVisibility = NewVisibility(false)

func mustAlbumTitle(s string) domain.AlbumTitle {
	t, err := domain.NewAlbumTitle(s)
	if err != nil {
		panic(err)
	}
	return t
}

type Visibility struct {
	public bool
}

func NewVisibility(public bool) Visibility {
	return Visibility{public: public}
}

func (v Visibility) Public() bool {
	return v.public
}

type Library struct {
	root RootAlbum
}

func NewLibrary(root RootAlbum) *Library {
	return &Library{root: root}
}

func (l *Library) GetRootAlbum(publicOnly bool) (domain.RootAlbum, error) {
	rootVis := visibilityOr(l.root.visibility)
	if publicOnly && !rootVis.Public() {
		return domain.RootAlbum{}, errors.Wrap(ErrAlbumNotFound, "root not visible")
	}

	children := buildChildren(l.root.albums, rootVis, publicOnly)

	var tracks []domain.Track
	if !publicOnly || rootVis.Public() {
		tracks = append(tracks, l.root.tracks...)
		SortTracks(tracks)
	}

	return domain.NewRootAlbum(l.root.thumbnail, children, tracks)
}

func (l *Library) GetAlbum(id domain.AlbumId, publicOnly bool) (domain.Album, error) {
	path, ok := l.findAlbumPath(id)
	if !ok {
		return domain.Album{}, errors.Wrapf(ErrAlbumNotFound, "album '%s'", id)
	}

	effective := visibilityOr(l.root.visibility)
	for _, a := range path {
		if a.visibility != nil {
			effective = *a.visibility
		}
	}

	if publicOnly && !effective.Public() {
		return domain.Album{}, errors.Wrapf(ErrAlbumNotFound, "album '%s' not visible", id)
	}

	target := path[len(path)-1]

	parents := make([]domain.ParentAlbum, 0, len(path))
	for _, a := range path {
		parents = append(parents, domain.NewParentAlbum(a.id, a.title))
	}

	children := buildChildren(target.albums, effective, publicOnly)
	tracks := append([]domain.Track(nil), target.tracks...)
	SortTracks(tracks)

	if len(children) == 0 && len(tracks) == 0 {
		return domain.Album{}, errors.Wrapf(ErrAlbumNotFound, "album '%s' has no visible content", id)
	}

	return domain.NewAlbum(target.id, target.title, target.thumbnail, parents, children, tracks)
}

func (l *Library) GetTrack(id domain.TrackId, publicOnly bool) (domain.Track, error) {
	rootVis := visibilityOr(l.root.visibility)
	for _, t := range l.root.tracks {
		if t.Id() == id {
			if publicOnly && !rootVis.Public() {
				return domain.Track{}, errors.Wrapf(ErrTrackNotFound, "track '%s' not visible", id)
			}
			return t, nil
		}
	}

	t, vis, ok := findTrackInAlbums(l.root.albums, id, rootVis)
	if !ok {
		return domain.Track{}, errors.Wrapf(ErrTrackNotFound, "track '%s'", id)
	}
	if publicOnly && !vis.Public() {
		return domain.Track{}, errors.Wrapf(ErrTrackNotFound, "track '%s' not visible", id)
	}
	return t, nil
}

func (l *Library) GetThumbnail(id domain.ThumbnailId, publicOnly bool) (domain.Thumbnail, error) {
	rootVis := visibilityOr(l.root.visibility)
	if t := l.root.thumbnail; t != nil && t.Id() == id {
		if publicOnly && !rootVis.Public() {
			return domain.Thumbnail{}, errors.Wrapf(ErrThumbnailNotFound, "thumbnail '%s' not visible", id)
		}
		return *t, nil
	}

	t, vis, ok := findThumbnailInAlbums(l.root.albums, id, rootVis)
	if !ok {
		return domain.Thumbnail{}, errors.Wrapf(ErrThumbnailNotFound, "thumbnail '%s'", id)
	}
	if publicOnly && !vis.Public() {
		return domain.Thumbnail{}, errors.Wrapf(ErrThumbnailNotFound, "thumbnail '%s' not visible", id)
	}
	return t, nil
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

type BasicAlbum struct {
	Path      []domain.AlbumId
	Title     domain.AlbumTitle
	Thumbnail *domain.Thumbnail
}

type SearchResult struct {
	Albums []BasicAlbum
	Tracks []SearchResultTrack
}

type SearchResultTrack struct {
	Track domain.Track
	Album BasicAlbum
}

// Search walks the whole tree, filtering by visibility cascade.
func (l *Library) Search(query string, publicOnly bool) (SearchResult, error) {
	var result SearchResult
	rootVis := visibilityOr(l.root.visibility)

	rootBasic := BasicAlbum{Title: rootAlbumTitle, Thumbnail: l.root.thumbnail}

	if !publicOnly || rootVis.Public() {
		for _, t := range l.root.tracks {
			if len(result.Tracks) >= maxSearchItems {
				break
			}
			if !containsStringCaseInsensitive(t.Title().String(), query) {
				continue
			}
			result.Tracks = append(result.Tracks, SearchResultTrack{Track: t, Album: rootBasic})
		}
	}

	searchAlbums(&result, l.root.albums, nil, rootVis, query, publicOnly)
	return result, nil
}

func searchAlbums(result *SearchResult, albums []Album, parentPath []domain.AlbumId, parentVis Visibility, query string, publicOnly bool) {
	for _, a := range albums {
		effective := parentVis
		if a.visibility != nil {
			effective = *a.visibility
		}
		if publicOnly && !effective.Public() {
			continue
		}

		path := make([]domain.AlbumId, 0, len(parentPath)+1)
		path = append(path, parentPath...)
		path = append(path, a.id)

		basic := BasicAlbum{Path: path, Title: a.title, Thumbnail: a.thumbnail}

		if len(result.Albums) < maxSearchItems && containsStringCaseInsensitive(a.title.String(), query) {
			result.Albums = append(result.Albums, basic)
		}
		for _, t := range a.tracks {
			if len(result.Tracks) >= maxSearchItems {
				break
			}
			if !containsStringCaseInsensitive(t.Title().String(), query) {
				continue
			}
			result.Tracks = append(result.Tracks, SearchResultTrack{Track: t, Album: basic})
		}

		searchAlbums(result, a.albums, path, effective, query, publicOnly)
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

func visibilityOr(v *Visibility) Visibility {
	if v == nil {
		return defaultVisibility
	}
	return *v
}

func buildChildren(albums []Album, parentVis Visibility, publicOnly bool) []domain.ChildAlbum {
	var out []domain.ChildAlbum
	for _, a := range albums {
		vis := parentVis
		if a.visibility != nil {
			vis = *a.visibility
		}
		if publicOnly && !vis.Public() {
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

func SortTracks(tracks []domain.Track) {
	sort.Slice(tracks, func(i, j int) bool {
		titleI := tracks[i].Title().String()
		titleJ := tracks[j].Title().String()

		fieldsI := strings.Fields(titleI)
		fieldsJ := strings.Fields(titleJ)

		if len(fieldsI) > 0 && len(fieldsJ) > 0 {
			f := func(r rune) bool { return !unicode.IsNumber(r) }
			numI, errI := strconv.Atoi(strings.TrimFunc(fieldsI[0], f))
			numJ, errJ := strconv.Atoi(strings.TrimFunc(fieldsJ[0], f))
			if errI == nil && errJ == nil {
				if numI == numJ {
					return strings.Join(fieldsI[1:], "") < strings.Join(fieldsJ[1:], "")
				}
				return numI < numJ
			}
		}
		return titleI < titleJ
	})
}

func containsStringCaseInsensitive(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
