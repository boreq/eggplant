package music

import (
	"context"
	"errors"

	"github.com/boreq/eggplant/adapters/music/scanner"
	"github.com/boreq/eggplant/domain"
)

var ErrForbidden = errors.New("forbidden")
var ErrNotFound = errors.New("not found")

type ThumbnailStore interface {
	GetConvertedFile(ctx context.Context, id domain.FileId) (ConvertedFile, error)
}

type TrackStore interface {
	GetConvertedFile(ctx context.Context, id domain.FileId) (ConvertedFile, error)
}

type Library interface {
	Browse(ids []domain.AlbumId, publicOnly bool) (domain.Album, error)
	Search(query string, publicOnly bool) (SearchResult, error)
	Apply(scan scanner.Album) error
}

type SearchResult struct {
	Albums []BasicAlbum
	Tracks []SearchResultTrack
}

type SearchResultTrack struct {
	Track domain.Track
	Album BasicAlbum
}

type BasicAlbum struct {
	Path      []domain.AlbumId
	Title     domain.AlbumTitle
	Thumbnail *domain.Thumbnail
}
