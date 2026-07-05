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
	"github.com/boreq/eggplant/domain/remote"
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

type Music struct {
	Thumbnail              *LoggingThumbnailHandler
	StartStreaming         *LoggingStartStreamingHandler
	StreamPlaylist         *LoggingStreamPlaylistHandler
	StreamInit             *LoggingStreamInitHandler
	StreamFragment         *LoggingStreamFragmentHandler
	KeepAliveStream        *LoggingKeepAliveStreamHandler
	GetRootAlbum           *LoggingGetRootAlbumHandler
	GetAlbum               *LoggingGetAlbumHandler
	RemoteGetAlbum         *LoggingRemoteGetAlbumHandler
	RemoteGetThumbnail     *LoggingRemoteGetThumbnailHandler
	RemoteGetTrackDuration *LoggingRemoteGetTrackDurationHandler
	RemoteStartStreaming   *LoggingRemoteStartStreamingHandler
	RemoteStreamPlaylist   *LoggingRemoteStreamPlaylistHandler
	RemoteStreamInit       *LoggingRemoteStreamInitHandler
	RemoteStreamFragment   *LoggingRemoteStreamFragmentHandler
	RemoteKeepAliveStream  *LoggingRemoteKeepAliveStreamHandler
	GetTrackDuration       *LoggingGetTrackDurationHandler
	Search                 *LoggingSearchHandler
	LoadLibrary            *LoggingLoadLibraryHandler
}

type TrackConverter interface {
	SetItems(items []TrackStoreItem)
	StartStream(ctx context.Context, fileId music.FileId, seekPosition *music.SeekPosition) (music.StreamId, error)
	GetPlaylist(fileId music.FileId, streamId music.StreamId) (Playlist, error)
	GetInit(fileId music.FileId, streamId music.StreamId) (ConvertedFile, error)
	GetFragment(fileId music.FileId, streamId music.StreamId, fragmentId music.FragmentId) (ConvertedFile, error)
	KeepAliveStream(fileId music.FileId, streamId music.StreamId) error
}

type TrackStoreItem struct {
	fileId music.FileId
	path   music.FilePath
}

func NewTrackStoreItem(fileId music.FileId, path music.FilePath) TrackStoreItem {
	return TrackStoreItem{fileId: fileId, path: path}
}

func (t TrackStoreItem) FileId() music.FileId {
	return t.fileId
}

func (t TrackStoreItem) Path() music.FilePath {
	return t.path
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

type TrackDurationStore interface {
	SetItems(items []TrackStoreItem)
	GetDuration(ctx context.Context, fileId music.FileId) (music.TrackDuration, error)
}

type AccessLoader interface {
	Load(file string) (library.Visibility, error)
}

type LibraryRepository interface {
	Get() (*library.Library, error)
	Save(library *library.Library)
}

type RemoteLibrary interface {
	GetRootAlbums(ctx context.Context) ([]music.RootAlbum, error)
	GetAlbum(ctx context.Context, instanceId remote.RemoteInstanceID, albumId music.AlbumId) (music.Album, error)
	GetThumbnail(ctx context.Context, instanceId remote.RemoteInstanceID, thumbnailId music.ThumbnailId) (io.ReadCloser, error)
	GetTrackDuration(ctx context.Context, instanceId remote.RemoteInstanceID, trackId music.TrackId) (music.TrackDuration, error)
	StartTrackStream(ctx context.Context, instanceId remote.RemoteInstanceID, trackId music.TrackId, seekPosition *music.RequestedSeekPosition) (music.StreamId, error)
	GetStreamPlaylist(ctx context.Context, instanceId remote.RemoteInstanceID, trackId music.TrackId, streamId music.StreamId) (io.ReadCloser, error)
	GetStreamInit(ctx context.Context, instanceId remote.RemoteInstanceID, trackId music.TrackId, streamId music.StreamId) (io.ReadCloser, error)
	GetStreamFragment(ctx context.Context, instanceId remote.RemoteInstanceID, trackId music.TrackId, streamId music.StreamId, fragmentId music.FragmentId) (io.ReadCloser, error)
	KeepAliveStream(ctx context.Context, instanceId remote.RemoteInstanceID, trackId music.TrackId, streamId music.StreamId) error
	Search(ctx context.Context, query string) (library.SearchResults, error)
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
