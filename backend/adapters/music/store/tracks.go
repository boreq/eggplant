package store

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/boreq/eggplant/application/music"
	"github.com/boreq/eggplant/domain"
	"github.com/boreq/eggplant/logging"
	"github.com/boreq/errors"
)

const (
	trackExtension = "ogg"
	trackDirectory = "tracks"
)

type TrackStore struct {
	*Store
}

func NewTrackStore(ctx context.Context, dataDir string) (*TrackStore, error) {
	log := logging.New("trackStore")
	converter := NewTrackConverter(dataDir)
	store, err := NewStore(ctx, log, converter)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a store")
	}
	return &TrackStore{Store: store}, nil
}

func (s *TrackStore) SetItems(items []music.TrackStoreItem) {
	converted := make([]Item, 0, len(items))
	for _, item := range items {
		converted = append(converted, Item{
			Id:   item.Id().String(),
			Path: item.Path().String(),
		})
	}
	s.Store.SetItems(converted)
}

func (s *TrackStore) GetConvertedFile(ctx context.Context, id domain.TrackId) (domain.ConvertedFile, error) {
	return s.Store.getConvertedFileForId(ctx, id.String())
}

type TrackConverter struct {
	dataDir string
	log     logging.Logger
}

func NewTrackConverter(dataDir string) *TrackConverter {
	converter := &TrackConverter{
		dataDir: dataDir,
		log:     logging.New("trackConverter"),
	}
	return converter
}

func (c *TrackConverter) Convert(item Item) error {
	outputPath := c.OutputFile(item.Id)
	tmpOutputPath := c.TemporaryOutputFile(item.Id)

	args := []string{
		"-y",
		"-i",
		item.Path,
		"-vn",
		"-c:a",
		"libopus",
		"-b:a",
		"96K",
		tmpOutputPath,
	}
	cmd := exec.Command("ffmpeg", args...)
	bufErr := &bytes.Buffer{}
	cmd.Stderr = bufErr
	c.log.Debug("converting", "command", cmd.String())
	if err := cmd.Run(); err != nil {
		c.log.Error("command error", "stderr", bufErr.String())
		return errors.Wrap(err, "ffmpeg execution failed")
	}

	if err := os.Rename(tmpOutputPath, outputPath); err != nil {
		return errors.Wrap(err, "move failed")
	}

	return nil
}

func (c *TrackConverter) OutputDirectory() string {
	return path.Join(c.dataDir, trackDirectory)
}

func (c *TrackConverter) OutputFile(id string) string {
	file := fmt.Sprintf("%s.%s", id, trackExtension)
	return path.Join(c.OutputDirectory(), file)
}

func (c *TrackConverter) TemporaryOutputFile(id string) string {
	file := fmt.Sprintf("_%s.%s", id, trackExtension)
	return path.Join(c.OutputDirectory(), file)
}
