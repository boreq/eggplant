package thumbnails

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/boreq/eggplant/application/music"
	musicdomain "github.com/boreq/eggplant/domain/music"
	"github.com/boreq/eggplant/internal/logging"
	"github.com/boreq/errors"
)

const thumbnailDirectory = "thumbnails"

type ThumbnailFormat interface {
	Extension() string
	Convert(inputPath string, outputPath string) error
}

type ThumbnailStore struct {
	*Store
}

func NewThumbnailStore(ctx context.Context, dataDir string, format ThumbnailFormat) (*ThumbnailStore, error) {
	log := logging.New("thumbnailStore")
	converter := NewThumbnailConverter(dataDir, format)
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

func (s *ThumbnailStore) GetConvertedFile(ctx context.Context, fileId musicdomain.FileId) (music.ConvertedFile, error) {
	return s.getConvertedFileForId(ctx, fileId.String())
}

func NewThumbnailConverter(dataDir string, format ThumbnailFormat) *ThumbnailConverter {
	converter := &ThumbnailConverter{
		dataDir: dataDir,
		format:  format,
		log:     logging.New("thumbnailConverter"),
	}
	return converter
}

type ThumbnailConverter struct {
	dataDir string
	format  ThumbnailFormat
	log     logging.Logger
}

func (c *ThumbnailConverter) Convert(item Item) error {
	outputPath := c.OutputFile(item.Id)
	tmpOutputPath := c.TemporaryOutputFile(item.Id)

	if err := os.MkdirAll(c.OutputDirectory(), 0755); err != nil {
		return errors.Wrap(err, "could not create the output directory")
	}

	if err := c.format.Convert(item.Path, tmpOutputPath); err != nil {
		return errors.Wrap(err, "conversion failed")
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
	file := fmt.Sprintf("%s.%s", id, c.format.Extension())
	return path.Join(c.OutputDirectory(), file)
}

func (c *ThumbnailConverter) TemporaryOutputFile(id string) string {
	file := fmt.Sprintf("_%s.%s", id, c.format.Extension())
	return path.Join(c.OutputDirectory(), file)
}
