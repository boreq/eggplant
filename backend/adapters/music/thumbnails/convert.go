package thumbnails

import (
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/boreq/errors"
	"github.com/nfnt/resize"
)

const thumbnailSize = 300

func decodeAndResize(inputPath string) (image.Image, error) {
	f, err := os.Open(inputPath)
	if err != nil {
		return nil, errors.Wrap(err, "could not open input file")
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, errors.Wrap(err, "decoding failed")
	}

	return resize.Resize(thumbnailSize, thumbnailSize, img, resize.Lanczos3), nil
}
