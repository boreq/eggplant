package store

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/boreq/eggplant/logging"
	"github.com/boreq/errors"
)

const (
	trackExtension = "ogg"
	trackDirectory = "tracks"
)

type TrackStore struct {
	*Store
	durationCache      map[string]time.Duration
	durationCacheMutex sync.Mutex
	converter          *TrackConverter
	log                logging.Logger
}

func NewTrackStore(ctx context.Context, dataDir string) (*TrackStore, error) {
	log := logging.New("trackStore")
	converter := NewTrackConverter(dataDir)
	store, err := NewStore(ctx, log, converter)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a store")
	}
	s := &TrackStore{
		Store:         store,
		durationCache: make(map[string]time.Duration),
		converter:     converter,
		log:           log,
	}
	return s, nil
}

func (s *TrackStore) GetDuration(id string) time.Duration {
	s.durationCacheMutex.Lock()
	defer s.durationCacheMutex.Unlock()

	item, ok := s.getItem(id)
	if !ok {
		return 0
	}

	if duration, ok := s.durationCache[id]; ok {
		return duration
	}

	duration, err := s.converter.checkDuration(item)
	if err != nil {
		s.log.Debug("duration could not be measured", "err", err)
		return 0
	}
	s.durationCache[id] = duration
	return duration
}

func (s *TrackStore) getItem(id string) (Item, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	i, ok := s.items[id]
	return i, ok
}

func (s *TrackStore) SetItems(items []Item) {
	s.cleanupDurationCache(items)
	s.Store.SetItems(items)
}

func (s *TrackStore) cleanupDurationCache(items []Item) {
	s.durationCacheMutex.Lock()
	defer s.durationCacheMutex.Unlock()

	existingItems := make(map[string]bool)
	for _, item := range items {
		existingItems[item.Id] = true
	}

	for id := range s.durationCache {
		if _, exists := existingItems[id]; !exists {
			delete(s.durationCache, id)
		}
	}
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

func (c *TrackConverter) checkDuration(item Item) (time.Duration, error) {
	filePath := item.Path

	// check if a file exists at all
	if _, err := os.Stat(filePath); err != nil {
		return 0, errors.Wrap(err, "stat failed")
	}

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
