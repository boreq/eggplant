package store

import (
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"os"
	"path"

	"github.com/boreq/eggplant/logging"
	"github.com/boreq/errors"
	"github.com/nfnt/resize"
)

const thumbnailSize = 200
const thumbnailExtension = "jpg"
const thumbnailDirectory = "thumbnails"

func NewThumbnailStore(dataDir string) (*Store, error) {
	log := logging.New("thumbnailStore")
	converter := NewThumbnailConverter(dataDir)
	return NewStore(log, converter)
}

func NewThumbnailConverter(dataDir string) *ThumbnailConverter {
	converter := &ThumbnailConverter{
		dataDir: dataDir,
		log:     logging.New("thumbnailConverter"),
	}
	return converter
}

type ThumbnailConverter struct {
	dataDir string
	log     logging.Logger
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

	options := &jpeg.Options{
		Quality: 95,
	}
	if err := jpeg.Encode(output, resized, options); err != nil {
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
	return path.Join(c.dataDir, thumbnailDirectory)
}

func (c *ThumbnailConverter) tmpOutputFile(id string) string {
	file := fmt.Sprintf("_%s.%s", id, thumbnailExtension)
	return path.Join(c.OutputDirectory(), file)
}
