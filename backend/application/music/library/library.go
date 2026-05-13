// Package library is responsible for orchestrating actions related to
// providing a navigable representation of the audio library.
package library

import (
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/boreq/eggplant/adapters/music/scanner"
	"github.com/boreq/eggplant/adapters/music/store"
	"github.com/boreq/eggplant/application/music"
	"github.com/boreq/eggplant/domain"
	"github.com/boreq/eggplant/logging"
	"github.com/boreq/errors"
)

const rootAlbumTitle = "Eggplant"

type TrackStore interface {
	SetItems(items []store.Item)
	GetDuration(id string) time.Duration
}

type ThumbnailStore interface {
	SetItems(items []store.Item)
}

type AccessLoader interface {
	Load(file string) (domain.Access, error)
}

type IdGenerator interface {
	AlbumId(parents []domain.AlbumId, title string) (domain.AlbumId, error)
	TrackId(parents []domain.AlbumId, title string) (domain.TrackId, error)
	FileId(path string) (domain.FileId, error)
}

// Library receives scanner updates, dispatches them to appropriate stores and
// builds a navigable representation of the music collection.
type Library struct {
	trackStore     TrackStore
	thumbnailStore ThumbnailStore
	accessLoader   AccessLoader
	idGenerator    IdGenerator
	root           *album
	mutex          sync.Mutex
	log            logging.Logger
}

// New creates an empty library. Apply scanner updates via Apply.
func New(
	trackStore TrackStore,
	thumbnailStore ThumbnailStore,
	accessLoader AccessLoader,
	idGenerator IdGenerator,
) (*Library, error) {
	l := &Library{
		trackStore:     trackStore,
		thumbnailStore: thumbnailStore,
		accessLoader:   accessLoader,
		idGenerator:    idGenerator,
		root:           newAlbum(rootAlbumTitle),
		log:            logging.New("library"),
	}
	return l, nil
}

// Browse lists the specified album. Provide a zero-length slice to list the
// root album.
func (l *Library) Browse(ids []domain.AlbumId, publicOnly bool) (domain.Album, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	a, err := l.getAlbum(ids)
	if err != nil {
		return domain.Album{}, errors.Wrap(err, "failed to get an album")
	}

	access, err := l.getAccess(ids)
	if err != nil {
		return domain.Album{}, errors.Wrap(err, "failed to get access")
	}

	parents, err := l.getParents(ids)
	if err != nil {
		return domain.Album{}, errors.Wrap(err, "failed to get parents")
	}

	title, err := domain.NewAlbumTitle(a.title)
	if err != nil {
		return domain.Album{}, errors.Wrap(err, "invalid album title")
	}

	var id domain.AlbumId
	if len(ids) > 0 {
		id = ids[len(ids)-1]
	}

	var childAlbums []domain.Album
	for childId, child := range a.albums {
		childAccess, err := l.getAccess(append(ids, childId))
		if err != nil {
			return domain.Album{}, errors.Wrap(err, "failed to get access")
		}

		if !canAccess(childAccess, publicOnly) {
			continue
		}

		childTitle, err := domain.NewAlbumTitle(child.title)
		if err != nil {
			return domain.Album{}, errors.Wrap(err, "invalid album title")
		}

		childAlbum, err := domain.NewAlbum(childId, childTitle, thumbnailFor(child), childAccess, nil, nil, nil)
		if err != nil {
			return domain.Album{}, errors.Wrap(err, "could not build child album")
		}
		childAlbums = append(childAlbums, childAlbum)
	}
	sortAlbums(childAlbums)

	var tracks []domain.Track
	if canAccess(access, publicOnly) {
		for trackId, t := range a.tracks {
			trackTitle, err := domain.NewTrackTitle(t.title)
			if err != nil {
				return domain.Album{}, errors.Wrap(err, "invalid track title")
			}
			tracks = append(tracks, domain.NewTrack(
				trackId,
				t.fileId,
				trackTitle,
				l.trackDuration(t.fileId),
			))
		}
		SortTracks(tracks)
	}

	album, err := domain.NewAlbum(id, title, thumbnailFor(a), access, parents, childAlbums, tracks)
	if err != nil {
		return domain.Album{}, errors.Wrap(err, "could not build album")
	}
	return album, nil
}

const maxSearchItems = 10

func (l *Library) Search(query string, publicOnly bool) (music.SearchResult, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	var result music.SearchResult

	if err := l.walk(
		func(parent *music.BasicAlbum, id domain.AlbumId, v album) error {
			if len(result.Albums) > maxSearchItems {
				return nil
			}

			if !containsStringCaseInsensitive(v.title, query) {
				return nil
			}

			var path []domain.AlbumId
			if parent != nil {
				path = append(path, parent.Path...)
			}
			path = append(path, id)

			ba, err := newBasicAlbum(path, v)
			if err != nil {
				return errors.Wrap(err, "could not build basic album")
			}
			result.Albums = append(result.Albums, ba)

			return nil
		},
		func(parents music.BasicAlbum, id domain.TrackId, v track) error {
			if len(result.Tracks) > maxSearchItems {
				return nil
			}

			if !containsStringCaseInsensitive(v.title, query) {
				return nil
			}

			srt, err := l.toSearchResultTrack(parents, id, v)
			if err != nil {
				return errors.Wrap(err, "could not build search result track")
			}
			result.Tracks = append(result.Tracks, srt)
			return nil
		},
		publicOnly,
	); err != nil {
		return music.SearchResult{}, errors.Wrap(err, "walk failed")
	}

	return result, nil
}

func (l *Library) toSearchResultTrack(album music.BasicAlbum, id domain.TrackId, v track) (music.SearchResultTrack, error) {
	title, err := domain.NewTrackTitle(v.title)
	if err != nil {
		return music.SearchResultTrack{}, errors.Wrap(err, "invalid track title")
	}
	return music.SearchResultTrack{
		Track: domain.NewTrack(id, v.fileId, title, l.trackDuration(v.fileId)),
		Album: album,
	}, nil
}

// trackDuration returns a non-nil pointer if the track's duration is known
// and positive; otherwise nil (the duration could not be determined).
func (l *Library) trackDuration(fileId domain.FileId) *domain.TrackDuration {
	d, err := domain.NewTrackDuration(l.trackStore.GetDuration(fileId.String()))
	if err != nil {
		return nil
	}
	return &d
}

func (l *Library) getParents(ids []domain.AlbumId) ([]domain.AlbumParent, error) {
	parents := make([]domain.AlbumParent, 0)
	for i := 0; i < len(ids); i++ {
		parentIds := ids[:i+1]
		dir, err := l.getAlbum(parentIds)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get a parent album")
		}
		title, err := domain.NewAlbumTitle(dir.title)
		if err != nil {
			return nil, errors.Wrap(err, "invalid album title")
		}
		parents = append(parents, domain.NewAlbumParent(parentIds[len(parentIds)-1], title))
	}
	return parents, nil
}

var defaultAccess = domain.NewAccess(false)

func (l *Library) getAccess(ids []domain.AlbumId) (domain.Access, error) {
	for i := len(ids); i >= 0; i-- {
		parentIds := ids[:i]
		album, err := l.getAlbum(parentIds)
		if err != nil {
			return domain.Access{}, errors.Wrap(err, "failed to get a parent album")
		}
		if album.access != nil {
			return *album.access, nil
		}
	}
	return defaultAccess, nil
}

// Apply replaces the library state with the contents of the given scan.
func (l *Library) Apply(album scanner.Album) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// update the track list
	l.root = newAlbum(rootAlbumTitle)
	if err := l.mergeAlbum(nil, l.root, album); err != nil {
		return errors.Wrap(err, "merge album failed")
	}

	// inform track store which files are available for conversion
	var tracks []store.Item
	if err := l.getTracks(&tracks, l.root); err != nil {
		return errors.Wrap(err, "preparing tracks failed")
	}
	l.trackStore.SetItems(tracks)

	// inform thumbnail store which files are available for conversion
	var thumbnails []store.Item
	if err := l.getThumbnails(&thumbnails, l.root); err != nil {
		return errors.Wrap(err, "preparing thumbnails failed")
	}
	l.thumbnailStore.SetItems(thumbnails)

	return nil
}

func (l *Library) mergeAlbum(parents []domain.AlbumId, target *album, album scanner.Album) error {
	if album.Thumbnail != "" {
		thumbnailId, err := l.idGenerator.FileId(album.Thumbnail)
		if err != nil {
			return errors.Wrap(err, "could not create a thumbnail id")
		}
		target.thumbnailPath = album.Thumbnail
		target.thumbnailId = &thumbnailId
	}

	if album.AccessFile != "" {
		acc, err := l.accessLoader.Load(album.AccessFile)
		if err != nil {
			return errors.Wrap(err, "could not load the access file")
		}
		target.access = &acc
	}

	for title, scannerTrack := range album.Tracks {
		id, track, err := l.toTrack(parents, title, scannerTrack)
		if err != nil {
			return errors.Wrap(err, "could not convert to a track")
		}
		target.tracks[id] = track
	}

	for title, scannerAlbum := range album.Albums {
		id, album, err := l.toAlbum(parents, title, *scannerAlbum)
		if err != nil {
			return errors.Wrap(err, "could not convert to an album")
		}
		target.albums[id] = album

		childParents := append(parents, id)
		if err := l.mergeAlbum(childParents, album, *scannerAlbum); err != nil {
			return err
		}
	}

	return nil
}

func (l *Library) getThumbnails(thumbnails *[]store.Item, current *album) error {
	if current.thumbnailPath != "" {
		thumbnail := store.Item{
			Id:   current.thumbnailId.String(),
			Path: current.thumbnailPath,
		}
		*thumbnails = append(*thumbnails, thumbnail)
	}

	for _, child := range current.albums {
		if err := l.getThumbnails(thumbnails, child); err != nil {
			return err
		}
	}

	return nil
}

func (l *Library) getTracks(tracks *[]store.Item, current *album) error {
	for _, track := range current.tracks {
		track := store.Item{
			Id:   track.fileId.String(),
			Path: track.path,
		}
		*tracks = append(*tracks, track)
	}

	for _, child := range current.albums {
		if err := l.getTracks(tracks, child); err != nil {
			return err
		}
	}

	return nil
}

func (l *Library) getAlbum(ids []domain.AlbumId) (*album, error) {
	var current *album = l.root
	for _, id := range ids {
		child, ok := current.albums[id]
		if !ok {
			return nil, errors.Wrapf(music.ErrNotFound, "album '%s' not found", id)
		}
		current = child
	}
	return current, nil
}

func (l *Library) newTrack(title string, path string) (track, error) {
	fileId, err := l.idGenerator.FileId(path)
	if err != nil {
		return track{}, errors.Wrap(err, "could not create a file id")
	}
	t := track{
		title:  title,
		path:   path,
		fileId: fileId,
	}
	return t, nil
}

func (l *Library) toTrack(parents []domain.AlbumId, title string, scannerTrack scanner.Track) (domain.TrackId, track, error) {
	id, err := l.idGenerator.TrackId(parents, title)
	if err != nil {
		return domain.TrackId{}, track{}, errors.Wrap(err, "could not create a track id")
	}
	t, err := l.newTrack(title, scannerTrack.Path)
	if err != nil {
		return domain.TrackId{}, track{}, errors.Wrap(err, "could not create a track")
	}
	return id, t, nil
}

func (l *Library) toAlbum(parents []domain.AlbumId, title string, scannerAlbum scanner.Album) (domain.AlbumId, *album, error) {
	id, err := l.idGenerator.AlbumId(parents, title)
	if err != nil {
		return domain.AlbumId{}, nil, errors.Wrap(err, "could not create an album id")
	}
	album := newAlbum(title)
	return id, album, nil
}

func canAccess(access domain.Access, publicOnly bool) bool {
	if publicOnly && !access.Public() {
		return false
	}
	return true
}

// thumbnailFor returns a domain.Thumbnail pointer for the internal album if it
// has a thumbnail set; otherwise nil.
func thumbnailFor(a *album) *domain.Thumbnail {
	if a.thumbnailId == nil {
		return nil
	}
	t := domain.NewThumbnail(*a.thumbnailId)
	return &t
}

type track struct {
	title  string
	path   string
	fileId domain.FileId
}

type album struct {
	title         string
	thumbnailPath string
	thumbnailId   *domain.FileId
	access        *domain.Access
	albums        map[domain.AlbumId]*album
	tracks        map[domain.TrackId]track
}

func newAlbum(title string) *album {
	return &album{
		title:  title,
		albums: make(map[domain.AlbumId]*album),
		tracks: make(map[domain.TrackId]track),
	}
}

func sortAlbums(albums []domain.Album) {
	sort.Slice(albums,
		func(i, j int) bool {
			return albums[i].Title().String() < albums[j].Title().String()
		},
	)
}

func SortTracks(tracks []domain.Track) {
	sort.Slice(tracks,
		func(i, j int) bool {
			titleI := tracks[i].Title().String()
			titleJ := tracks[j].Title().String()

			fieldsI := strings.Fields(titleI)
			fieldsJ := strings.Fields(titleJ)

			if len(fieldsI) > 0 && len(fieldsJ) > 0 {
				f := func(r rune) bool {
					return !unicode.IsNumber(r)
				}

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
		},
	)
}

func containsStringCaseInsensitive(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
