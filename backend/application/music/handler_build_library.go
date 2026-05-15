package music

import (
	"github.com/boreq/eggplant/domain"
	"github.com/boreq/eggplant/domain/library"
	scannerdomain "github.com/boreq/eggplant/domain/scanner"
	"github.com/boreq/errors"
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

func (h *BuildLibraryHandler) Execute(scan scannerdomain.FoundRootAlbum) error {
	b := &libraryBuilder{
		accessLoader: h.accessLoader,
		durations:    h.durations,
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

type libraryBuilder struct {
	accessLoader   AccessLoader
	durations      TrackDurations
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

		duration, err := b.durations.GetDuration(path.String())
		if err != nil {
			return nil, errors.Wrapf(err, "could not measure duration of '%s'", path)
		}

		id, err := domain.NewTrackId(parents, title)
		if err != nil {
			return nil, errors.Wrapf(err, "could not generate track id for '%s'", title)
		}

		out = append(out, domain.NewTrack(id, title, duration))
		b.trackItems = append(b.trackItems, NewTrackStoreItem(id, path))
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

	b.thumbnailItems = append(b.thumbnailItems, NewThumbnailStoreItem(thumbnailId, *file))

	t := domain.NewThumbnail(thumbnailId)
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
