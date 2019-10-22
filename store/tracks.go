package store

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/boreq/eggplant/logging"
	"github.com/pkg/errors"
)

func NewTrackStore(cacheDir string) (*TrackStore, error) {
	converter := NewTrackConverter(cacheDir)
	store, err := NewStore(converter)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a store")
	}
	s := &TrackStore{
		Store:     store,
		converter: converter,
		log:       logging.New("trackStore"),
	}
	return s, nil
}

type TrackStore struct {
	*Store
	converter *TrackConverter
	log       logging.Logger
}

func (s *TrackStore) GetDuration(id string) time.Duration {
	duration, err := s.converter.checkDuration(id)
	if err != nil {
		s.log.Error("duration could not be measured", "err", err)
	}
	return duration
}

func NewTrackConverter(cacheDir string) *TrackConverter {
	converter := &TrackConverter{
		cacheDir: cacheDir,
		log:      logging.New("trackConverter"),
	}
	return converter
}

type TrackConverter struct {
	cacheDir string
	log      logging.Logger
}

func (c *TrackConverter) Convert(item Item) error {
	outputPath := c.OutputFile(item.Id)
	tmpOutputPath := c.tmpOutputFile(item.Id)

	if err := makeDirectory(outputPath); err != nil {
		return errors.Wrap(err, "could not create output directory")
	}

	args := []string{
		"-y",
		"-i",
		item.Path,
		"-vn",
		"-c:a",
		"libopus",
		"-b:a",
		"96K",
		"-threads",
		"4",
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

func (c *TrackConverter) checkDuration(id string) (time.Duration, error) {
	filePath := c.OutputFile(id)

	args := []string{
		"-v",
		"error",
		"-show_entries",
		"format=duration",
		"-of",
		"default=noprint_wrappers=1:nokey=1",
		filePath,
	}
	cmd := exec.Command("ffprobe", args...)
	bufErr := &bytes.Buffer{}
	cmd.Stderr = bufErr
	c.log.Debug("checking duration", "command", cmd.String())
	output, err := cmd.Output()
	if err != nil {
		c.log.Error("command error", "stderr", bufErr.String())
		return 0, errors.Wrap(err, "ffprobe execution failed")
	}

	normalized := strings.TrimSpace(string(output)) + "s"
	duration, err := time.ParseDuration(normalized)
	if err != nil {
		return 0, errors.Wrap(err, "could not parse the duration")
	}
	return duration, nil
}

const trackExtension = "ogg"
const trackDirectory = "tracks"

func (c *TrackConverter) OutputFile(id string) string {
	dir := path.Join(c.cacheDir, trackDirectory)
	file := fmt.Sprintf("%s.%s", id, trackExtension)
	return path.Join(dir, file)
}

func (c *TrackConverter) tmpOutputFile(id string) string {
	dir := path.Join(c.cacheDir, trackDirectory)
	file := fmt.Sprintf("_%s.%s", id, trackExtension)
	return path.Join(dir, file)
}
