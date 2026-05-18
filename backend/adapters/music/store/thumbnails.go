package store

import (
	"context"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path"

	"github.com/boreq/eggplant/application/music"
	"github.com/boreq/eggplant/domain"
	"github.com/boreq/eggplant/logging"
	"github.com/boreq/errors"
	"github.com/chai2010/webp"
	"github.com/nfnt/resize"
)

const thumbnailSize = 300
const thumbnailQuality = 90
const thumbnailExtension = "webp"
const thumbnailDirectory = "thumbnails"

type ThumbnailStore struct {
	*Store
}

func NewThumbnailStore(ctx context.Context, dataDir string) (*ThumbnailStore, error) {
	log := logging.New("thumbnailStore")
	converter := NewThumbnailConverter(dataDir)
	store, err := NewStore(ctx, log, converter)
	if err != nil {
		return nil, err
	}
	return &ThumbnailStore{Store: store}, nil
}

func (s *ThumbnailStore) SetItems(items []music.ThumbnailStoreItem) {
	converted := make([]Item, 0, len(items))
	for _, item := range items {
		converted = append(converted, Item{
			Id:   item.FileId().String(),
			Path: item.Path().String(),
		})
	}
	s.Store.SetItems(converted)
}

func (s *ThumbnailStore) GetConvertedFile(ctx context.Context, fileId domain.FileId) (music.ConvertedFile, error) {
	return s.getConvertedFileForId(ctx, fileId.String())
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
	tmpOutputPath := c.TemporaryOutputFile(item.Id)

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

	if err := webp.Encode(output, resized, &webp.Options{Quality: thumbnailQuality}); err != nil {
		return errors.Wrap(err, "encoding failed")
	}

	if err := os.Rename(tmpOutputPath, outputPath); err != nil {
		return errors.Wrap(err, "move failed")
	}

	return nil
}

func (c *ThumbnailConverter) OutputDirectory() string {
	return path.Join(c.dataDir, thumbnailDirectory)
}

func (c *ThumbnailConverter) OutputFile(id string) string {
	file := fmt.Sprintf("%s.%s", id, thumbnailExtension)
	return path.Join(c.OutputDirectory(), file)
}

func (c *ThumbnailConverter) TemporaryOutputFile(id string) string {
	file := fmt.Sprintf("_%s.%s", id, thumbnailExtension)
	return path.Join(c.OutputDirectory(), file)
}
