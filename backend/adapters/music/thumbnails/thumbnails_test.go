package thumbnails_test

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/boreq/eggplant/adapters/music/thumbnails"
	"github.com/stretchr/testify/require"
)

func TestThumbnailConverterCreatesOutputDirectory(t *testing.T) {
	dataDir := t.TempDir()

	inputPath := filepath.Join(dataDir, "input.png")
	createTestImage(t, inputPath)

	converter := thumbnails.NewThumbnailConverter(dataDir, thumbnails.NewGen2BrainWebpFormat(thumbnails.DefaultGen2BrainWebpQuality))

	_, err := os.Stat(converter.OutputDirectory())
	require.True(t, os.IsNotExist(err))

	item := thumbnails.Item{
		Id:   "someid",
		Path: inputPath,
	}

	err = converter.Convert(item)
	require.NoError(t, err)

	_, err = os.Stat(converter.OutputFile(item.Id))
	require.NoError(t, err)
}

func createTestImage(t *testing.T, path string) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for x := 0; x < 10; x++ {
		for y := 0; y < 10; y++ {
			img.Set(x, y, color.RGBA{R: 255, A: 255})
		}
	}

	f, err := os.Create(path)
	require.NoError(t, err)
	defer f.Close()

	err = png.Encode(f, img)
	require.NoError(t, err)
}
