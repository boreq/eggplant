//go:generate go run github.com/boreq/eggplant/internal/cmd/genhandlers -dir .

package music

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/boreq/eggplant/domain/music"
	"github.com/boreq/eggplant/domain/music/hls"
	"github.com/boreq/eggplant/domain/music/library"
)

var (
	ErrLibraryNotReady        = errors.New("library not ready")
	ErrTooManyOpenStreams     = errors.New("too many open streams")
	ErrStreamNotFound         = errors.New("stream not found")
	ErrStreamTrackMismatch    = errors.New("stream does not belong to this track")
	ErrStreamPlaylistNotFound = errors.New("stream playlist not found")
	ErrStreamInitNotFound     = errors.New("stream init not found")
	ErrStreamFragmentNotFound = errors.New("stream fragment not found")
)

type TrackConverter interface {
	SetItems(items []TrackStoreItem)
	StartStream(ctx context.Context, fileId music.FileId, seekPosition *music.SeekPosition) (music.StreamId, error)
	GetPlaylist(fileId music.FileId, streamId music.StreamId) (Playlist, error)
	GetInit(fileId music.FileId, streamId music.StreamId) (ConvertedFile, error)
	GetFragment(fileId music.FileId, streamId music.StreamId, fragmentId music.FragmentId) (ConvertedFile, error)
	KeepAliveStream(fileId music.FileId, streamId music.StreamId) error
}

type TrackStoreItem struct {
	fileId   music.FileId
	path     music.FilePath
	duration music.TrackDuration
}

func NewTrackStoreItem(fileId music.FileId, path music.FilePath, duration music.TrackDuration) TrackStoreItem {
	return TrackStoreItem{fileId: fileId, path: path, duration: duration}
}

func (t TrackStoreItem) FileId() music.FileId {
	return t.fileId
}

func (t TrackStoreItem) Path() music.FilePath {
	return t.path
}

func (t TrackStoreItem) Duration() music.TrackDuration {
	return t.duration
}

type ThumbnailStore interface {
	SetItems(items []ThumbnailStoreItem)
	GetConvertedFile(ctx context.Context, fileId music.FileId) (ConvertedFile, error)
}

type ThumbnailStoreItem struct {
	fileId music.FileId
	path   music.FilePath
}

func NewThumbnailStoreItem(fileId music.FileId, path music.FilePath) ThumbnailStoreItem {
	return ThumbnailStoreItem{fileId: fileId, path: path}
}

func (t ThumbnailStoreItem) FileId() music.FileId {
	return t.fileId
}

func (t ThumbnailStoreItem) Path() music.FilePath {
	return t.path
}

type TrackDurations interface {
	GetDuration(ctx context.Context, path string) (music.TrackDuration, error)
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
	ErrStreamNotFound,
	ErrStreamTrackMismatch,
	ErrStreamPlaylistNotFound,
	ErrStreamInitNotFound,
	ErrStreamFragmentNotFound,
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
