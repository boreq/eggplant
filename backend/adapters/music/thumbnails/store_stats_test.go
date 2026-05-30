package thumbnails_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/boreq/eggplant/adapters/music/thumbnails"
	"github.com/boreq/eggplant/internal/logging"
	"github.com/stretchr/testify/require"
)

type fakeConverter struct {
	outputDir string
}

func (f fakeConverter) OutputFile(id string) string {
	return filepath.Join(f.outputDir, id)
}

func (f fakeConverter) TemporaryOutputFile(id string) string {
	return filepath.Join(f.outputDir, id+".tmp")
}

func (f fakeConverter) OutputDirectory() string {
	return f.outputDir
}

func (f fakeConverter) Convert(item thumbnails.Item) error {
	return nil
}

func TestStoreGetStatsSkipsMissingFiles(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	converter := fakeConverter{outputDir: filepath.Join(t.TempDir(), "converted")}

	s, err := thumbnails.NewStore(ctx, logging.New("test"), converter)
	require.NoError(t, err)

	content := []byte("thumbnail data")
	existing := filepath.Join(t.TempDir(), "existing.jpg")
	require.NoError(t, os.WriteFile(existing, content, 0600))

	missing := filepath.Join(t.TempDir(), "missing.jpg")

	s.SetItems([]thumbnails.Item{
		{Id: "existing", Path: existing},
		{Id: "missing", Path: missing},
	})

	stats, err := s.GetStats()
	require.NoError(t, err)
	require.Equal(t, int64(2), stats.AllItems)
	require.Equal(t, int64(len(content)), stats.OriginalSize)
	require.Equal(t, int64(0), stats.ConvertedItems)
}
