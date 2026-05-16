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
	GetPlaylist(ctx context.Context, fileId domain.FileId) (domain.ConvertedFile, error)
	GetInit(ctx context.Context, fileId domain.FileId) (domain.ConvertedFile, error)
	GetFragment(ctx context.Context, fileId domain.FileId, fragmentId domain.TrackFragmentId) (domain.ConvertedFile, error)
}

type TrackStoreItem struct {
	fileId domain.FileId
	path   domain.FilePath
}

func NewTrackStoreItem(fileId domain.FileId, path domain.FilePath) TrackStoreItem {
	return TrackStoreItem{fileId: fileId, path: path}
}

func (t TrackStoreItem) FileId() domain.FileId {
	return t.fileId
}

func (t TrackStoreItem) Path() domain.FilePath {
	return t.path
}

type ThumbnailStore interface {
	SetItems(items []ThumbnailStoreItem)
	GetConvertedFile(ctx context.Context, fileId domain.FileId) (domain.ConvertedFile, error)
}

type ThumbnailStoreItem struct {
	fileId domain.FileId
	path   domain.FilePath
}

func NewThumbnailStoreItem(fileId domain.FileId, path domain.FilePath) ThumbnailStoreItem {
	return ThumbnailStoreItem{fileId: fileId, path: path}
}

func (t ThumbnailStoreItem) FileId() domain.FileId {
	return t.fileId
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
