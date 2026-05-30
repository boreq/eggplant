package componenttest

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/boreq/eggplant/entrypoints/http/openapi"
	"github.com/boreq/eggplant/internal/config"
	"github.com/boreq/eggplant/internal/wire"
	"github.com/stretchr/testify/require"
)

func TestServiceInitialSetup(t *testing.T) {
	ts := newTestService(t)
	ctx := context.Background()

	statsResp, err := ts.client.GetStatsWithResponse(ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, statsResp.StatusCode())
	require.NotNil(t, statsResp.JSON200)
	require.Equal(t, 0, statsResp.JSON200.Users)

	registerResp, err := ts.client.RegisterInitialWithResponse(ctx, openapi.RegisterInitialJSONRequestBody{
		Username: "admin",
		Password: "password",
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, registerResp.StatusCode())

	loginResp, err := ts.client.LoginWithResponse(ctx, openapi.LoginJSONRequestBody{
		Username: "admin",
		Password: "password",
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, loginResp.StatusCode())
	require.NotNil(t, loginResp.JSON200)
	require.NotEmpty(t, loginResp.JSON200.Token)

	meResp, err := ts.client.GetCurrentUserWithResponse(ctx, authedAs(loginResp.JSON200.Token))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, meResp.StatusCode())
	require.NotNil(t, meResp.JSON200)
	require.Equal(t, "admin", meResp.JSON200.Username)
	require.True(t, meResp.JSON200.Administrator)

	registerResp, err = ts.client.RegisterInitialWithResponse(ctx, openapi.RegisterInitialJSONRequestBody{
		Username: "other",
		Password: "password",
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusInternalServerError, registerResp.StatusCode())
}

func TestServiceAuthFlow(t *testing.T) {
	ts := newTestService(t)
	ctx := context.Background()

	token := registerAdminAndLogin(t, ts)
	admin := authedAs(token)

	t.Run("current user is the administrator", func(t *testing.T) {
		resp, err := ts.client.GetCurrentUserWithResponse(ctx, admin)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())
		require.NotNil(t, resp.JSON200)
		require.Equal(t, "admin", resp.JSON200.Username)
		require.True(t, resp.JSON200.Administrator)
	})

	t.Run("current user is unauthorized without a token", func(t *testing.T) {
		resp, err := ts.client.GetCurrentUserWithResponse(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode())
	})

	t.Run("invite, register and list users", func(t *testing.T) {
		inviteResp, err := ts.client.CreateInvitationWithResponse(ctx, admin)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, inviteResp.StatusCode())
		require.NotNil(t, inviteResp.JSON200)

		registerResp, err := ts.client.RegisterWithResponse(ctx, openapi.RegisterJSONRequestBody{
			Username: "bob",
			Password: "password",
			Token:    inviteResp.JSON200.Token,
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, registerResp.StatusCode())

		usersResp, err := ts.client.GetUsersWithResponse(ctx, admin)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, usersResp.StatusCode())
		require.NotNil(t, usersResp.JSON200)
		require.Len(t, *usersResp.JSON200, 2)
	})

	t.Run("listing users is forbidden for anonymous callers", func(t *testing.T) {
		resp, err := ts.client.GetUsersWithResponse(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusForbidden, resp.StatusCode())
	})

	t.Run("remove user", func(t *testing.T) {
		resp, err := ts.client.RemoveUserWithResponse(ctx, "bob", admin)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())

		usersResp, err := ts.client.GetUsersWithResponse(ctx, admin)
		require.NoError(t, err)
		require.NotNil(t, usersResp.JSON200)
		require.Len(t, *usersResp.JSON200, 1)
	})

	t.Run("logout", func(t *testing.T) {
		resp, err := ts.client.LogoutWithResponse(ctx, admin)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())
	})
}

func TestServiceMeta(t *testing.T) {
	ts := newTestService(t)
	ctx := context.Background()

	token := registerAdminAndLogin(t, ts)

	t.Run("version", func(t *testing.T) {
		resp, err := ts.client.GetVersionWithResponse(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())
		require.NotNil(t, resp.JSON200)
		require.NotEmpty(t, resp.JSON200.Version)
	})

	t.Run("stats reflect the registered user", func(t *testing.T) {
		resp, err := ts.client.GetStatsWithResponse(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())
		require.NotNil(t, resp.JSON200)
		require.Equal(t, 1, resp.JSON200.Users)
		require.Equal(t, int64(1), resp.JSON200.Tracks.NumberOfTracks)
	})

	_ = token
}

func TestServiceBrowseAndStream(t *testing.T) {
	ts := newTestService(t)
	ctx := context.Background()

	token := registerAdminAndLogin(t, ts)
	admin := authedAs(token)

	// Browse the root: it must contain the fixture album.
	rootResp, err := ts.client.BrowseWithResponse(ctx, admin)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rootResp.StatusCode())
	require.NotNil(t, rootResp.JSON200)
	require.Len(t, rootResp.JSON200.Albums, 1)

	albumID := rootResp.JSON200.Albums[0].Id

	// Browse the album: it must contain the track and a thumbnail.
	albumResp, err := ts.client.BrowseAlbumWithResponse(ctx, albumID, admin)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, albumResp.StatusCode())
	require.NotNil(t, albumResp.JSON200)
	require.Len(t, albumResp.JSON200.Tracks, 1)
	require.NotNil(t, albumResp.JSON200.Thumbnail)

	trackID := albumResp.JSON200.Tracks[0].Id
	thumbnailID := albumResp.JSON200.Thumbnail.Id

	t.Run("browsing is denied for anonymous callers", func(t *testing.T) {
		resp, err := ts.client.BrowseWithResponse(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())
		require.NotNil(t, resp.JSON200)
		require.Empty(t, resp.JSON200.Albums)
	})

	t.Run("search finds the track", func(t *testing.T) {
		resp, err := ts.client.SearchWithResponse(ctx, &openapi.SearchParams{Query: "Song"}, admin)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())
		require.NotNil(t, resp.JSON200)
		require.NotEmpty(t, resp.JSON200.Tracks)
	})

	t.Run("track duration", func(t *testing.T) {
		resp, err := ts.client.GetTrackDurationWithResponse(ctx, trackID, admin)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())
		require.NotNil(t, resp.JSON200)
		require.Greater(t, resp.JSON200.Duration, 0.0)
	})

	t.Run("thumbnail is converted to webp", func(t *testing.T) {
		resp, err := ts.client.GetThumbnailWithResponse(ctx, thumbnailID, admin)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())
		require.Equal(t, "image/webp", resp.HTTPResponse.Header.Get("Content-Type"))
		require.NotEmpty(t, resp.Body)
	})

	t.Run("hls streaming pipeline", func(t *testing.T) {
		startResp, err := ts.client.StartTrackStreamWithResponse(ctx, trackID, &openapi.StartTrackStreamParams{}, admin)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, startResp.StatusCode())
		require.NotNil(t, startResp.JSON200)
		streamID := startResp.JSON200.StreamId

		playlistResp, err := ts.client.GetStreamPlaylistWithResponse(ctx, trackID, streamID, admin)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, playlistResp.StatusCode())
		require.NotEmpty(t, playlistResp.Body)

		initResp, err := ts.client.GetStreamInitWithResponse(ctx, trackID, streamID, admin)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, initResp.StatusCode())
		require.NotEmpty(t, initResp.Body)

		fragmentResp, err := ts.client.GetStreamFragmentWithResponse(ctx, trackID, streamID, 0, admin)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, fragmentResp.StatusCode())
		require.NotEmpty(t, fragmentResp.Body)

		keepAliveResp, err := ts.client.KeepAliveStreamWithResponse(ctx, trackID, streamID, admin)
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, keepAliveResp.StatusCode())
	})
}

func TestServiceDocs(t *testing.T) {
	ts := newTestService(t)

	t.Run("redoc page is served", func(t *testing.T) {
		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, ts.baseURL+"/api", nil)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Contains(t, resp.Header.Get("Content-Type"), "text/html")

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Contains(t, string(body), "<redoc")
	})

	t.Run("openapi spec file is served", func(t *testing.T) {
		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, ts.baseURL+"/api/openapi.yaml", nil)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Contains(t, string(body), "openapi: 3.0.3")
	})
}

type testService struct {
	baseURL string
	client  *openapi.ClientWithResponses
}

func newTestService(t *testing.T) *testService {
	requireBinary(t, "ffmpeg")
	requireBinary(t, "ffprobe")

	musicDir := t.TempDir()
	writeFixtureLibrary(t, musicDir)

	conf := config.Default()
	conf.MusicDirectory = musicDir
	conf.DataDirectory = t.TempDir()
	conf.CacheDirectory = t.TempDir()
	conf.ServeAddress = "127.0.0.1:0"

	ctx, cancel := context.WithCancel(context.Background())

	svc, err := wire.BuildTestHTTPService(ctx, conf)
	require.NoError(t, err)

	// Scan the fixture library synchronously so the music endpoints have data.
	require.NoError(t, svc.App.Music.LoadLibrary.Execute(ctx))

	server := httptest.NewServer(svc.Handler)

	t.Cleanup(func() {
		server.Close()
		cancel()
		_ = svc.DB.Close()
	})

	client, err := openapi.NewClientWithResponses(server.URL)
	require.NoError(t, err)

	return &testService{baseURL: server.URL, client: client}
}

func authedAs(token string) openapi.RequestEditorFn {
	return func(ctx context.Context, req *http.Request) error {
		req.AddCookie(&http.Cookie{Name: "auth-token", Value: token})
		return nil
	}
}

func registerAdminAndLogin(t *testing.T, ts *testService) string {
	ctx := context.Background()

	registerResp, err := ts.client.RegisterInitialWithResponse(ctx, openapi.RegisterInitialJSONRequestBody{
		Username: "admin",
		Password: "password",
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, registerResp.StatusCode())

	loginResp, err := ts.client.LoginWithResponse(ctx, openapi.LoginJSONRequestBody{
		Username: "admin",
		Password: "password",
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, loginResp.StatusCode())
	require.NotNil(t, loginResp.JSON200)
	require.NotEmpty(t, loginResp.JSON200.Token)

	return loginResp.JSON200.Token
}

func requireBinary(t *testing.T, name string) {
	if _, err := exec.LookPath(name); err != nil {
		t.Skipf("%s is not available", name)
	}
}

// writeFixtureLibrary synthesises one album (one track + thumbnail) with ffmpeg.
func writeFixtureLibrary(t *testing.T, musicDir string) {
	albumDir := filepath.Join(musicDir, "Cool Album")
	require.NoError(t, os.MkdirAll(albumDir, 0755))

	runFFmpeg(t,
		"-f", "lavfi", "-i", "sine=frequency=440:duration=5",
		filepath.Join(albumDir, "Song One.flac"),
	)
	runFFmpeg(t,
		"-f", "lavfi", "-i", "color=c=red:s=64x64:d=1", "-frames:v", "1",
		filepath.Join(albumDir, "thumbnail.jpg"),
	)
}

func runFFmpeg(t *testing.T, args ...string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ffmpeg", append([]string{"-nostdin", "-y"}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("ffmpeg failed: %v\n%s", err, out)
	}
}
