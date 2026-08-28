package tracks_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/boreq/eggplant/adapters/music/tracks"
	"github.com/boreq/eggplant/application/music"
	musicdomain "github.com/boreq/eggplant/domain/music"
	"github.com/stretchr/testify/require"
)

func TestConverterGetStatsSkipsMissingFiles(t *testing.T) {
	ctx := t.Context()

	dataDir := t.TempDir()

	converter, err := tracks.NewConverter(ctx, dataDir)
	require.NoError(t, err)

	content := []byte("hello world")
	existing := filepath.Join(t.TempDir(), "existing.mp3")
	require.NoError(t, os.WriteFile(existing, content, 0600))

	missing := filepath.Join(t.TempDir(), "missing.mp3")

	converter.SetItems([]music.TrackStoreItem{
		newTrackStoreItem(t, existing),
		newTrackStoreItem(t, missing),
	})

	stats, err := converter.GetStats()
	require.NoError(t, err)
	require.Equal(t, int64(2), stats.NumberOfTracks)
	require.Equal(t, int64(len(content)), stats.SizeOfTracks)
}

func newTrackStoreItem(t *testing.T, path string) music.TrackStoreItem {
	t.Helper()
	filePath, err := musicdomain.NewFilePath(path)
	require.NoError(t, err)
	fileId, err := musicdomain.NewFileId(filePath)
	require.NoError(t, err)
	return music.NewTrackStoreItem(fileId, filePath)
}
