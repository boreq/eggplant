package thumbnails

import (
	"os"

	"github.com/boreq/errors"
	gen2webp "github.com/gen2brain/webp"
)

const DefaultGen2BrainWebpQuality = 90

type Gen2BrainWebpFormat struct {
	quality int
}

func NewGen2BrainWebpFormat(quality int) *Gen2BrainWebpFormat {
	return &Gen2BrainWebpFormat{quality: quality}
}

func (f *Gen2BrainWebpFormat) Extension() string {
	return "webp"
}

func (f *Gen2BrainWebpFormat) Convert(inputPath string, outputPath string) error {
	img, err := decodeAndResize(inputPath)
	if err != nil {
		return err
	}

	out, err := os.Create(outputPath)
	if err != nil {
		return errors.Wrap(err, "could not create output file")
	}
	defer out.Close()

	if err := gen2webp.Encode(out, img, gen2webp.Options{Quality: f.quality}); err != nil {
		return errors.Wrap(err, "webp encoding failed")
	}
	return nil
}
