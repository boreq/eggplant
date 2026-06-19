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

	bounds := img.Bounds()
	var width, height uint
	if bounds.Dx() < bounds.Dy() {
		width = thumbnailSize
	} else {
		height = thumbnailSize
	}

	return resize.Resize(width, height, img, resize.Lanczos3), nil
}
