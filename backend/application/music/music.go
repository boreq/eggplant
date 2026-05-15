package music

import (
	"context"
	"errors"

	"github.com/boreq/eggplant/domain"
	"github.com/boreq/eggplant/domain/library"
)

var (
	ErrForbidden       = errors.New("forbidden")
	ErrLibraryNotReady = errors.New("library not ready")
)

type TrackStore interface {
	SetItems(items []TrackStoreItem)
	GetConvertedFile(ctx context.Context, id domain.TrackId) (domain.ConvertedFile, error)
}

type TrackStoreItem struct {
	id   domain.TrackId
	path domain.FilePath
}

func NewTrackStoreItem(id domain.TrackId, path domain.FilePath) TrackStoreItem {
	return TrackStoreItem{id: id, path: path}
}

func (t TrackStoreItem) Id() domain.TrackId {
	return t.id
}

func (t TrackStoreItem) Path() domain.FilePath {
	return t.path
}

type ThumbnailStore interface {
	SetItems(items []ThumbnailStoreItem)
	GetConvertedFile(ctx context.Context, id domain.ThumbnailId) (domain.ConvertedFile, error)
}

type ThumbnailStoreItem struct {
	id   domain.ThumbnailId
	path domain.FilePath
}

func NewThumbnailStoreItem(id domain.ThumbnailId, path domain.FilePath) ThumbnailStoreItem {
	return ThumbnailStoreItem{id: id, path: path}
}

func (t ThumbnailStoreItem) Id() domain.ThumbnailId {
	return t.id
}

func (t ThumbnailStoreItem) Path() domain.FilePath {
	return t.path
}

type TrackDurations interface {
	GetDuration(path string) (domain.TrackDuration, error)
}

type AccessLoader interface {
	Load(file string) (library.Visibility, error)
}

type LibraryRepository interface {
	Get() (*library.Library, error)
	Save(library *library.Library)
}
