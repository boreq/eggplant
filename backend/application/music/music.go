//go:generate go run github.com/boreq/eggplant/internal/cmd/genhandlers -dir .

package music

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/boreq/eggplant/domain"
	"github.com/boreq/eggplant/domain/hls"
	"github.com/boreq/eggplant/domain/library"
)

var (
	ErrLibraryNotReady    = errors.New("library not ready")
	ErrTooManyOpenStreams = errors.New("too many open streams")
)

type TrackConverter interface {
	SetItems(items []TrackStoreItem)
	StartStream(ctx context.Context, fileId domain.FileId, seekPosition *domain.SeekPosition) (domain.StreamId, error)
	GetPlaylist(fileId domain.FileId, streamId domain.StreamId) (Playlist, error)
	GetInit(fileId domain.FileId, streamId domain.StreamId) (ConvertedFile, error)
	GetFragment(fileId domain.FileId, streamId domain.StreamId, fragmentId domain.FragmentId) (ConvertedFile, error)
}

type TrackStoreItem struct {
	fileId   domain.FileId
	path     domain.FilePath
	duration domain.TrackDuration
}

func NewTrackStoreItem(fileId domain.FileId, path domain.FilePath, duration domain.TrackDuration) TrackStoreItem {
	return TrackStoreItem{fileId: fileId, path: path, duration: duration}
}

func (t TrackStoreItem) FileId() domain.FileId {
	return t.fileId
}

func (t TrackStoreItem) Path() domain.FilePath {
	return t.path
}

func (t TrackStoreItem) Duration() domain.TrackDuration {
	return t.duration
}

type ThumbnailStore interface {
	SetItems(items []ThumbnailStoreItem)
	GetConvertedFile(ctx context.Context, fileId domain.FileId) (ConvertedFile, error)
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
	GetDuration(ctx context.Context, path string) (domain.TrackDuration, error)
}

type AccessLoader interface {
	Load(file string) (library.Visibility, error)
}

type LibraryRepository interface {
	Get() (*library.Library, error)
	Save(library *library.Library)
}

type Playlist struct {
	Playlist hls.Playlist
	Modtime  time.Time
}

type ConvertedFile struct {
	// Modtime is used to figure out if the content has changed.
	Modtime time.Time

	// Content must be closed by the caller.
	Content io.ReadSeekCloser
}

var nonLoggableErrors = []error{
	ErrLibraryNotReady,
	ErrTooManyOpenStreams,
	library.ErrAlbumNotFound,
	library.ErrTrackNotFound,
	library.ErrThumbnailNotFound,
}

func isNonLoggableError(err error) bool {
	for _, target := range nonLoggableErrors {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}
