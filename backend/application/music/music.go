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
	thumbnail              *LoggingThumbnailHandler
	startStreaming         *LoggingStartStreamingHandler
	streamPlaylist         *LoggingStreamPlaylistHandler
	streamInit             *LoggingStreamInitHandler
	streamFragment         *LoggingStreamFragmentHandler
	keepAliveStream        *LoggingKeepAliveStreamHandler
	getRootAlbum           *LoggingGetRootAlbumHandler
	getAlbum               *LoggingGetAlbumHandler
	remoteGetRootAlbum     *LoggingRemoteGetRootAlbumHandler
	remoteGetAlbum         *LoggingRemoteGetAlbumHandler
	remoteGetThumbnail     *LoggingRemoteGetThumbnailHandler
	remoteGetTrackDuration *LoggingRemoteGetTrackDurationHandler
	remoteStartStreaming   *LoggingRemoteStartStreamingHandler
	remoteStreamPlaylist   *LoggingRemoteStreamPlaylistHandler
	remoteStreamInit       *LoggingRemoteStreamInitHandler
	remoteStreamFragment   *LoggingRemoteStreamFragmentHandler
	remoteKeepAliveStream  *LoggingRemoteKeepAliveStreamHandler
	getTrackDuration       *LoggingGetTrackDurationHandler
	search                 *LoggingSearchHandler
	loadLibrary            *LoggingLoadLibraryHandler
}

func NewMusic(
	thumbnail *LoggingThumbnailHandler,
	startStreaming *LoggingStartStreamingHandler,
	streamPlaylist *LoggingStreamPlaylistHandler,
	streamInit *LoggingStreamInitHandler,
	streamFragment *LoggingStreamFragmentHandler,
	keepAliveStream *LoggingKeepAliveStreamHandler,
	getRootAlbum *LoggingGetRootAlbumHandler,
	getAlbum *LoggingGetAlbumHandler,
	remoteGetRootAlbum *LoggingRemoteGetRootAlbumHandler,
	remoteGetAlbum *LoggingRemoteGetAlbumHandler,
	remoteGetThumbnail *LoggingRemoteGetThumbnailHandler,
	remoteGetTrackDuration *LoggingRemoteGetTrackDurationHandler,
	remoteStartStreaming *LoggingRemoteStartStreamingHandler,
	remoteStreamPlaylist *LoggingRemoteStreamPlaylistHandler,
	remoteStreamInit *LoggingRemoteStreamInitHandler,
	remoteStreamFragment *LoggingRemoteStreamFragmentHandler,
	remoteKeepAliveStream *LoggingRemoteKeepAliveStreamHandler,
	getTrackDuration *LoggingGetTrackDurationHandler,
	search *LoggingSearchHandler,
	loadLibrary *LoggingLoadLibraryHandler,
) *Music {
	return &Music{
		thumbnail:              thumbnail,
		startStreaming:         startStreaming,
		streamPlaylist:         streamPlaylist,
		streamInit:             streamInit,
		streamFragment:         streamFragment,
		keepAliveStream:        keepAliveStream,
		getRootAlbum:           getRootAlbum,
		getAlbum:               getAlbum,
		remoteGetRootAlbum:     remoteGetRootAlbum,
		remoteGetAlbum:         remoteGetAlbum,
		remoteGetThumbnail:     remoteGetThumbnail,
		remoteGetTrackDuration: remoteGetTrackDuration,
		remoteStartStreaming:   remoteStartStreaming,
		remoteStreamPlaylist:   remoteStreamPlaylist,
		remoteStreamInit:       remoteStreamInit,
		remoteStreamFragment:   remoteStreamFragment,
		remoteKeepAliveStream:  remoteKeepAliveStream,
		getTrackDuration:       getTrackDuration,
		search:                 search,
		loadLibrary:            loadLibrary,
	}
}

func (m Music) Thumbnail() *LoggingThumbnailHandler {
	return m.thumbnail
}

func (m Music) StartStreaming() *LoggingStartStreamingHandler {
	return m.startStreaming
}

func (m Music) StreamPlaylist() *LoggingStreamPlaylistHandler {
	return m.streamPlaylist
}

func (m Music) StreamInit() *LoggingStreamInitHandler {
	return m.streamInit
}

func (m Music) StreamFragment() *LoggingStreamFragmentHandler {
	return m.streamFragment
}

func (m Music) KeepAliveStream() *LoggingKeepAliveStreamHandler {
	return m.keepAliveStream
}

func (m Music) GetRootAlbum() *LoggingGetRootAlbumHandler {
	return m.getRootAlbum
}

func (m Music) GetAlbum() *LoggingGetAlbumHandler {
	return m.getAlbum
}

func (m Music) RemoteGetRootAlbum() *LoggingRemoteGetRootAlbumHandler {
	return m.remoteGetRootAlbum
}

func (m Music) RemoteGetAlbum() *LoggingRemoteGetAlbumHandler {
	return m.remoteGetAlbum
}

func (m Music) RemoteGetThumbnail() *LoggingRemoteGetThumbnailHandler {
	return m.remoteGetThumbnail
}

func (m Music) RemoteGetTrackDuration() *LoggingRemoteGetTrackDurationHandler {
	return m.remoteGetTrackDuration
}

func (m Music) RemoteStartStreaming() *LoggingRemoteStartStreamingHandler {
	return m.remoteStartStreaming
}

func (m Music) RemoteStreamPlaylist() *LoggingRemoteStreamPlaylistHandler {
	return m.remoteStreamPlaylist
}

func (m Music) RemoteStreamInit() *LoggingRemoteStreamInitHandler {
	return m.remoteStreamInit
}

func (m Music) RemoteStreamFragment() *LoggingRemoteStreamFragmentHandler {
	return m.remoteStreamFragment
}

func (m Music) RemoteKeepAliveStream() *LoggingRemoteKeepAliveStreamHandler {
	return m.remoteKeepAliveStream
}

func (m Music) GetTrackDuration() *LoggingGetTrackDurationHandler {
	return m.getTrackDuration
}

func (m Music) Search() *LoggingSearchHandler {
	return m.search
}

func (m Music) LoadLibrary() *LoggingLoadLibraryHandler {
	return m.loadLibrary
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
	GetRootAlbum(ctx context.Context, instanceId remote.RemoteInstanceID) (music.RootAlbum, error)
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
	playlist hls.Playlist
	modtime  time.Time
}

func NewPlaylist(playlist hls.Playlist, modtime time.Time) Playlist {
	return Playlist{
		playlist: playlist,
		modtime:  modtime,
	}
}

func (p Playlist) Playlist() hls.Playlist {
	return p.playlist
}

func (p Playlist) Modtime() time.Time {
	return p.modtime
}

type ConvertedFile struct {
	// Modtime is used to figure out if the content has changed.
	modtime time.Time

	// Content must be closed by the caller.
	content io.ReadSeekCloser
}

func NewConvertedFile(modtime time.Time, content io.ReadSeekCloser) ConvertedFile {
	return ConvertedFile{
		modtime: modtime,
		content: content,
	}
}

func (f ConvertedFile) Modtime() time.Time {
	return f.modtime
}

func (f ConvertedFile) Content() io.ReadSeekCloser {
	return f.content
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
