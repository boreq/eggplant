package store

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"os"
	"path"

	"github.com/boreq/eggplant/logging"
	"github.com/nfnt/resize"
	"github.com/pkg/errors"
)

const thumbnailSize = 200
const thumbnailExtension = "png"
const thumbnailDirectory = "thumbnails"

func NewThumbnailStore(cacheDir string) (*Store, error) {
	log := logging.New("thumbnailStore")
	converter := NewThumbnailConverter(cacheDir)
	return NewStore(log, converter)
}

func NewThumbnailConverter(cacheDir string) *ThumbnailConverter {
	converter := &ThumbnailConverter{
		cacheDir: cacheDir,
		log:      logging.New("thumbnailConverter"),
	}
	return converter
}

type ThumbnailConverter struct {
	cacheDir string
	log      logging.Logger
}

func (c *ThumbnailConverter) Convert(item Item) error {
	outputPath := c.OutputFile(item.Id)
	tmpOutputPath := c.tmpOutputFile(item.Id)

	if err := makeDirectory(outputPath); err != nil {
		return errors.Wrap(err, "could not create the output directory")
	}

	f, err := os.Open(item.Path)
	if err != nil {
		return errors.Wrap(err, "could not open the input file")
	}
	defer f.Close()

	output, err := os.Create(tmpOutputPath)
	if err != nil {
		return errors.Wrap(err, "could not create an output file")
	}
	defer output.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return errors.Wrap(err, "decoding failed")
	}

	resized := resize.Resize(thumbnailSize, thumbnailSize, img, resize.Lanczos3)

	encoder := png.Encoder{CompressionLevel: png.BestCompression}
	if err := encoder.Encode(output, resized); err != nil {
		return errors.Wrap(err, "encoding failed")
	}

	if err := os.Rename(tmpOutputPath, outputPath); err != nil {
		return errors.Wrap(err, "move failed")
	}

	return nil
}

func (c *ThumbnailConverter) OutputFile(id string) string {
	file := fmt.Sprintf("%s.%s", id, thumbnailExtension)
	return path.Join(c.OutputDirectory(), file)
}

func (c *ThumbnailConverter) OutputDirectory() string {
	return path.Join(c.cacheDir, thumbnailDirectory)
}

func (c *ThumbnailConverter) tmpOutputFile(id string) string {
	file := fmt.Sprintf("_%s.%s", id, thumbnailExtension)
	return path.Join(c.OutputDirectory(), file)
}
