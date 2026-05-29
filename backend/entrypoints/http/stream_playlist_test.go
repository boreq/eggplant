package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/boreq/eggplant/application"
	"github.com/boreq/eggplant/application/auth"
	appmusic "github.com/boreq/eggplant/application/music"
	musicdomain "github.com/boreq/eggplant/domain/music"
	"github.com/boreq/eggplant/domain/music/hls"
	"github.com/boreq/eggplant/domain/music/library"
	httpentrypoint "github.com/boreq/eggplant/entrypoints/http"
	"github.com/boreq/eggplant/internal/logging"
	"github.com/stretchr/testify/require"
)

// Issue #83: ffmpeg rewrites the playlist several times per second while
// converting, each time with more segments. If the response is cacheable, a
// client that fetched an early, short playlist replays it forever (hls.js
// stalls with bufferStalledError) instead of picking up the later segments.
// The fix makes the response uncacheable: no Last-Modified (so no conditional
// 304) and Cache-Control: no-store.
func TestStreamPlaylistResponseIsNotCacheable(t *testing.T) {
	h := newTestHandler(t)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, playlistRequest(t))

	require.Equal(t, http.StatusOK, rec.Code)
	require.Empty(t, rec.Header().Get("Last-Modified"))
	require.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
}

func playlistRequest(t *testing.T) *http.Request {
	t.Helper()
	streamId, err := musicdomain.NewStreamId()
	require.NoError(t, err)
	return httptest.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/api/track/"+testTrackId(t).String()+"/stream/"+streamId.String()+"/playlist",
		nil,
	)
}

func newTestHandler(t *testing.T) *httpentrypoint.Handler {
	t.Helper()

	title, err := musicdomain.NewTrackTitle("track")
	require.NoError(t, err)
	path, err := musicdomain.NewFilePath("/tracks/track.flac")
	require.NoError(t, err)
	fileId, err := musicdomain.NewFileId(path)
	require.NoError(t, err)
	track := musicdomain.NewTrack(testTrackId(t), fileId, title)

	visibility := library.NewVisibility(true)
	root, err := library.NewRootAlbum(nil, &visibility, nil, []musicdomain.Track{track})
	require.NoError(t, err)

	inner := appmusic.NewStreamPlaylistHandler(
		staticLibraryRepository{library.NewLibrary(root)},
		fakeTrackConverter{playlist: testPlaylist(t)},
	)
	app := &application.Application{
		Music: application.Music{
			StreamPlaylist: appmusic.NewLoggingStreamPlaylistHandler(inner, logging.New("test")),
		},
	}

	h, err := httpentrypoint.NewHandler(app, anonymousAuthProvider{})
	require.NoError(t, err)
	return h
}

func testTrackId(t *testing.T) musicdomain.TrackId {
	t.Helper()
	title, err := musicdomain.NewTrackTitle("track")
	require.NoError(t, err)
	id, err := musicdomain.NewTrackId(nil, title)
	require.NoError(t, err)
	return id
}

func testPlaylist(t *testing.T) appmusic.Playlist {
	t.Helper()
	version, err := hls.NewVersion(7)
	require.NoError(t, err)
	target, err := hls.NewTargetDuration(4 * time.Second)
	require.NoError(t, err)
	seq, err := hls.NewMediaSequence(0)
	require.NoError(t, err)
	mapURI, err := hls.NewMapURI("init")
	require.NoError(t, err)
	segDuration, err := hls.NewSegmentDuration(4 * time.Second)
	require.NoError(t, err)
	segURI, err := hls.NewSegmentURI("fragment/0")
	require.NoError(t, err)
	p, err := hls.NewPlaylist(
		version, target, seq, hls.PlaylistTypeEvent, mapURI,
		[]hls.Segment{hls.NewSegment(segDuration, segURI)},
		true,
	)
	require.NoError(t, err)
	return appmusic.Playlist{Playlist: p}
}

type staticLibraryRepository struct {
	lib *library.Library
}

func (r staticLibraryRepository) Get() (*library.Library, error) { return r.lib, nil }
func (r staticLibraryRepository) Save(*library.Library)          {}

// fakeTrackConverter embeds the interface so only GetPlaylist needs a body; the
// other methods are never called by the playlist handler.
type fakeTrackConverter struct {
	appmusic.TrackConverter
	playlist appmusic.Playlist
}

func (c fakeTrackConverter) GetPlaylist(musicdomain.FileId, musicdomain.StreamId) (appmusic.Playlist, error) {
	return c.playlist, nil
}

type anonymousAuthProvider struct{}

func (anonymousAuthProvider) Get(*http.Request) (httpentrypoint.AccessContext, error) {
	return httpentrypoint.NewAccessContext(auth.NewAnonymousAccessContext()), nil
}
