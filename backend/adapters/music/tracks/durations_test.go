package tracks_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/boreq/eggplant/adapters/music/tracks"
	"github.com/boreq/eggplant/application/music"
	musicdomain "github.com/boreq/eggplant/domain/music"
	"github.com/stretchr/testify/require"
)

func TestDurationStore_ConcurrentRequestsForSameFileProbeOnce(t *testing.T) {
	ctx := t.Context()

	d, err := musicdomain.NewTrackDuration(42 * time.Second)
	require.NoError(t, err)

	checker := &fakeChecker{duration: d, gate: make(chan struct{})}
	store := tracks.NewDurationStore(ctx, checker)

	fileId := newFileId(t, "/music/song.flac")
	store.SetItems([]music.TrackStoreItem{trackStoreItem(t, fileId, "/music/song.flac")})

	const callers = 25
	var wg sync.WaitGroup
	results := make([]musicdomain.TrackDuration, callers)
	errs := make([]error, callers)
	for i := range callers {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			results[i], errs[i] = store.GetDuration(ctx, fileId)
		}(i)
	}

	// Let every caller reach the in-flight join before the single probe returns.
	time.Sleep(50 * time.Millisecond)
	close(checker.gate)
	wg.Wait()

	for i := range callers {
		require.NoErrorf(t, errs[i], "caller %d", i)
		require.Equal(t, d, results[i])
	}
	require.Equal(t, int64(1), checker.calls.Load(), "expected a single deduped probe")

	// A subsequent request is served from cache without another probe.
	got, err := store.GetDuration(ctx, fileId)
	require.NoError(t, err)
	require.Equal(t, d, got)
	require.Equal(t, int64(1), checker.calls.Load())
}

func TestDurationStore_UnknownFileId(t *testing.T) {
	ctx := t.Context()

	store := tracks.NewDurationStore(ctx, &fakeChecker{})

	_, err := store.GetDuration(ctx, newFileId(t, "/music/missing.flac"))
	require.Error(t, err)
}

type fakeChecker struct {
	calls    atomic.Int64
	duration musicdomain.TrackDuration
	gate     chan struct{} // closed to release blocked probes
}

func (c *fakeChecker) GetDuration(ctx context.Context, path string) (musicdomain.TrackDuration, error) {
	c.calls.Add(1)
	if c.gate != nil {
		select {
		case <-c.gate:
		case <-ctx.Done():
			return musicdomain.TrackDuration{}, ctx.Err()
		}
	}
	return c.duration, nil
}

func newFileId(t *testing.T, path string) musicdomain.FileId {
	t.Helper()
	p, err := musicdomain.NewFilePath(path)
	require.NoError(t, err)
	id, err := musicdomain.NewFileId(p)
	require.NoError(t, err)
	return id
}

func trackStoreItem(t *testing.T, fileId musicdomain.FileId, path string) music.TrackStoreItem {
	t.Helper()
	p, err := musicdomain.NewFilePath(path)
	require.NoError(t, err)
	return music.NewTrackStoreItem(fileId, p)
}
