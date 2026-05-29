package music

import (
	"context"

	"github.com/boreq/eggplant/domain/music"
	"github.com/boreq/eggplant/domain/music/library"
	scannerdomain "github.com/boreq/eggplant/domain/music/scanner"
	"github.com/boreq/eggplant/domain/music/titleparser"
	"github.com/boreq/errors"
)

type Scanner interface {
	Scan() (scannerdomain.FoundRootAlbum, error)
}

type LoadLibraryHandler struct {
	repo           LibraryRepository
	scanner        Scanner
	trackStore     TrackConverter
	thumbnailStore ThumbnailStore
	accessLoader   AccessLoader
	durations      TrackDurationStore
}

func NewLoadLibraryHandler(
	repo LibraryRepository,
	scanner Scanner,
	trackStore TrackConverter,
	thumbnailStore ThumbnailStore,
	accessLoader AccessLoader,
	durations TrackDurationStore,
) *LoadLibraryHandler {
	return &LoadLibraryHandler{
		repo:           repo,
		scanner:        scanner,
		trackStore:     trackStore,
		thumbnailStore: thumbnailStore,
		accessLoader:   accessLoader,
		durations:      durations,
	}
}

func (h *LoadLibraryHandler) Execute(ctx context.Context) error {
	scan, err := h.scanner.Scan()
	if err != nil {
		return errors.Wrap(err, "scan failed")
	}

	b := &libraryBuilder{
		accessLoader: h.accessLoader,
	}

	root, err := b.buildRoot(scan)
	if err != nil {
		return errors.Wrap(err, "could not build library root")
	}

	h.trackStore.SetItems(b.trackItems)
	h.thumbnailStore.SetItems(b.thumbnailItems)
	h.durations.SetItems(b.trackItems)
	h.repo.Save(library.NewLibrary(root))
	return nil
}

type libraryBuilder struct {
	accessLoader   AccessLoader
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

func (b *libraryBuilder) buildAlbums(parents []music.AlbumId, src map[music.AlbumTitle]scannerdomain.FoundAlbum) ([]library.Album, error) {
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

func (b *libraryBuilder) buildAlbum(parents []music.AlbumId, title music.AlbumTitle, scan scannerdomain.FoundAlbum) (library.Album, error) {
	id, err := music.NewAlbumId(parents, title)
	if err != nil {
		return library.Album{}, errors.Wrap(err, "could not generate album id")
	}

	visibility, err := b.loadVisibility(scan.AccessFile())
	if err != nil {
		return library.Album{}, err
	}

	childParents := append(append([]music.AlbumId(nil), parents...), id)

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

func (b *libraryBuilder) buildTracks(parents []music.AlbumId, src map[music.TrackTitle]scannerdomain.FoundTrack) ([]music.Track, error) {
	var out []music.Track
	for title, ft := range src {
		path := ft.Path()

		id, err := music.NewTrackId(parents, title)
		if err != nil {
			return nil, errors.Wrapf(err, "could not generate track id for '%s'", title)
		}

		fileId, err := music.NewFileId(path)
		if err != nil {
			return nil, errors.Wrapf(err, "could not generate file id for '%s'", path)
		}

		out = append(out, music.NewTrack(id, fileId, title))
		b.trackItems = append(b.trackItems, NewTrackStoreItem(fileId, path))
	}
	return tryAddingTrackNumbers(out), nil
}

// tryAddingTrackNumbers only annotates tracks when every track in the album yields a parsed number,
// so we can reliably tell that the user most likely wants track numbers to be detected for this album.
func tryAddingTrackNumbers(tracks []music.Track) []music.Track {
	annotated := make([]music.Track, 0, len(tracks))
	for _, t := range tracks {
		parsed, err := titleparser.Parse(t.Title())
		if err != nil || parsed.Number() == nil {
			return tracks
		}
		annotated = append(annotated, music.NewTrackWithNumber(t.Id(), t.FileId(), *parsed.Number(), parsed.Title()))
	}
	return annotated
}

func (b *libraryBuilder) buildThumbnail(file *music.FilePath, parents []music.AlbumId) (*music.Thumbnail, error) {
	if file == nil {
		return nil, nil
	}

	name, err := music.NewFileNameFromFilePath(*file)
	if err != nil {
		return nil, errors.Wrap(err, "could not extract thumbnail filename")
	}

	thumbnailId, err := music.NewThumbnailId(parents, name)
	if err != nil {
		return nil, errors.Wrap(err, "could not generate thumbnail id")
	}

	fileId, err := music.NewFileId(*file)
	if err != nil {
		return nil, errors.Wrapf(err, "could not generate file id for '%s'", *file)
	}

	b.thumbnailItems = append(b.thumbnailItems, NewThumbnailStoreItem(fileId, *file))

	t := music.NewThumbnail(thumbnailId, fileId)
	return &t, nil
}

func (b *libraryBuilder) loadVisibility(file *music.FilePath) (*library.Visibility, error) {
	if file == nil {
		return nil, nil
	}
	v, err := b.accessLoader.Load(file.String())
	if err != nil {
		return nil, errors.Wrap(err, "could not load access file")
	}
	return &v, nil
}
