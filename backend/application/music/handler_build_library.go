package music

import (
	"context"
	"runtime"
	"sync"

	"github.com/boreq/eggplant/domain"
	"github.com/boreq/eggplant/domain/library"
	scannerdomain "github.com/boreq/eggplant/domain/scanner"
	"github.com/boreq/errors"
	"golang.org/x/sync/errgroup"
)

type BuildLibraryHandler struct {
	repo           LibraryRepository
	trackStore     TrackStore
	thumbnailStore ThumbnailStore
	accessLoader   AccessLoader
	durations      TrackDurations
}

func NewBuildLibraryHandler(
	repo LibraryRepository,
	trackStore TrackStore,
	thumbnailStore ThumbnailStore,
	accessLoader AccessLoader,
	durations TrackDurations,
) *BuildLibraryHandler {
	return &BuildLibraryHandler{
		repo:           repo,
		trackStore:     trackStore,
		thumbnailStore: thumbnailStore,
		accessLoader:   accessLoader,
		durations:      durations,
	}
}

func (h *BuildLibraryHandler) Execute(ctx context.Context, scan scannerdomain.FoundRootAlbum) error {
	pathDurations, err := h.probeDurations(ctx, scan)
	if err != nil {
		return errors.Wrap(err, "could not probe track durations")
	}

	b := &libraryBuilder{
		accessLoader:  h.accessLoader,
		pathDurations: pathDurations,
	}

	root, err := b.buildRoot(scan)
	if err != nil {
		return errors.Wrap(err, "could not build library root")
	}

	h.trackStore.SetItems(b.trackItems)
	h.thumbnailStore.SetItems(b.thumbnailItems)
	h.repo.Save(library.NewLibrary(root))
	return nil
}

func (h *BuildLibraryHandler) probeDurations(ctx context.Context, scan scannerdomain.FoundRootAlbum) (map[string]domain.TrackDuration, error) {
	paths := uniqueTrackPaths(scan)
	out := make(map[string]domain.TrackDuration, len(paths))
	var mu sync.Mutex

	var g errgroup.Group
	g.SetLimit(runtime.NumCPU())
	for _, p := range paths {
		g.Go(func() error {
			d, err := h.durations.GetDuration(ctx, p)
			if err != nil {
				return errors.Wrapf(err, "could not measure duration of '%s'", p)
			}
			mu.Lock()
			out[p] = d
			mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return out, nil
}

func uniqueTrackPaths(scan scannerdomain.FoundRootAlbum) []string {
	seen := map[string]struct{}{}
	var walk func(albums map[domain.AlbumTitle]scannerdomain.FoundAlbum, tracks map[domain.TrackTitle]scannerdomain.FoundTrack)
	walk = func(albums map[domain.AlbumTitle]scannerdomain.FoundAlbum, tracks map[domain.TrackTitle]scannerdomain.FoundTrack) {
		for _, t := range tracks {
			seen[t.Path().String()] = struct{}{}
		}
		for _, a := range albums {
			walk(a.Albums(), a.Tracks())
		}
	}
	walk(scan.Albums(), scan.Tracks())

	out := make([]string, 0, len(seen))
	for p := range seen {
		out = append(out, p)
	}
	return out
}

type libraryBuilder struct {
	accessLoader   AccessLoader
	pathDurations  map[string]domain.TrackDuration
	trackItems     []TrackStoreItem
	thumbnailItems []ThumbnailStoreItem
}

func (b *libraryBuilder) buildRoot(scan scannerdomain.FoundRootAlbum) (library.RootAlbum, error) {
	visibility, err := b.loadVisibility(scan.AccessFile())
	if err != nil {
		return library.RootAlbum{}, err
	}

	thumbnail, err := b.buildThumbnail(scan.ThumbnailFile(), nil)
	if err != nil {
		return library.RootAlbum{}, err
	}

	tracks, err := b.buildTracks(nil, scan.Tracks())
	if err != nil {
		return library.RootAlbum{}, err
	}

	albums, err := b.buildAlbums(nil, scan.Albums())
	if err != nil {
		return library.RootAlbum{}, err
	}

	return library.NewRootAlbum(thumbnail, visibility, albums, tracks)
}

func (b *libraryBuilder) buildAlbums(parents []domain.AlbumId, src map[domain.AlbumTitle]scannerdomain.FoundAlbum) ([]library.Album, error) {
	var out []library.Album
	for title, fa := range src {
		a, err := b.buildAlbum(parents, title, fa)
		if err != nil {
			return nil, errors.Wrapf(err, "could not build album '%s'", title)
		}
		out = append(out, a)
	}
	return out, nil
}

func (b *libraryBuilder) buildAlbum(parents []domain.AlbumId, title domain.AlbumTitle, scan scannerdomain.FoundAlbum) (library.Album, error) {
	id, err := domain.NewAlbumId(parents, title)
	if err != nil {
		return library.Album{}, errors.Wrap(err, "could not generate album id")
	}

	visibility, err := b.loadVisibility(scan.AccessFile())
	if err != nil {
		return library.Album{}, err
	}

	childParents := append(append([]domain.AlbumId(nil), parents...), id)

	thumbnail, err := b.buildThumbnail(scan.ThumbnailFile(), childParents)
	if err != nil {
		return library.Album{}, err
	}

	tracks, err := b.buildTracks(childParents, scan.Tracks())
	if err != nil {
		return library.Album{}, err
	}

	albums, err := b.buildAlbums(childParents, scan.Albums())
	if err != nil {
		return library.Album{}, err
	}

	album, err := library.NewAlbum(id, title, thumbnail, visibility, albums, tracks)
	if err != nil {
		return library.Album{}, errors.Wrap(err, "could not create album")
	}
	return album, nil
}

func (b *libraryBuilder) buildTracks(parents []domain.AlbumId, src map[domain.TrackTitle]scannerdomain.FoundTrack) ([]domain.Track, error) {
	var out []domain.Track
	for title, ft := range src {
		path := ft.Path()

		duration, ok := b.pathDurations[path.String()]
		if !ok {
			return nil, errors.New("missing duration for '" + path.String() + "'")
		}

		id, err := domain.NewTrackId(parents, title)
		if err != nil {
			return nil, errors.Wrapf(err, "could not generate track id for '%s'", title)
		}

		fileId, err := domain.NewFileId(path)
		if err != nil {
			return nil, errors.Wrapf(err, "could not generate file id for '%s'", path)
		}

		out = append(out, domain.NewTrack(id, fileId, title, duration))
		b.trackItems = append(b.trackItems, NewTrackStoreItem(fileId, path))
	}
	return out, nil
}

func (b *libraryBuilder) buildThumbnail(file *domain.FilePath, parents []domain.AlbumId) (*domain.Thumbnail, error) {
	if file == nil {
		return nil, nil
	}

	name, err := domain.NewFileNameFromFilePath(*file)
	if err != nil {
		return nil, errors.Wrap(err, "could not extract thumbnail filename")
	}

	thumbnailId, err := domain.NewThumbnailId(parents, name)
	if err != nil {
		return nil, errors.Wrap(err, "could not generate thumbnail id")
	}

	fileId, err := domain.NewFileId(*file)
	if err != nil {
		return nil, errors.Wrapf(err, "could not generate file id for '%s'", *file)
	}

	b.thumbnailItems = append(b.thumbnailItems, NewThumbnailStoreItem(fileId, *file))

	t := domain.NewThumbnail(thumbnailId, fileId)
	return &t, nil
}

func (b *libraryBuilder) loadVisibility(file *domain.FilePath) (*library.Visibility, error) {
	if file == nil {
		return nil, nil
	}
	v, err := b.accessLoader.Load(file.String())
	if err != nil {
		return nil, errors.Wrap(err, "could not load access file")
	}
	return &v, nil
}
