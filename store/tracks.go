package store

import (
	"bytes"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/boreq/eggplant/logging"
	"github.com/pkg/errors"
)

type Track struct {
	Id   string
	Path string
}

func NewTrackStore(cacheDir string) (*TrackStore, error) {
	s := &TrackStore{
		cacheDir: cacheDir,
		ch:       make(chan []Track),
		log:      logging.New("trackStore"),
	}
	go s.receiveTracks()
	go s.processTracks()
	return s, nil
}

type TrackStore struct {
	cacheDir  string
	ch        chan []Track
	tracks    []Track
	tracksSet bool
	mutex     sync.Mutex
	log       logging.Logger
}

func (s *TrackStore) SetTracks(tracks []Track) {
	s.ch <- tracks
}

func (s *TrackStore) ServeFile(w http.ResponseWriter, r *http.Request, id string) {
	http.ServeFile(w, r, s.filePath(id))
}

func (s *TrackStore) GetDuration(id string) time.Duration {
	duration, err := s.checkDuration(id)
	if err != nil {
		s.log.Error("duration could not be measured", "err", err)
	}
	return duration
}

func (s *TrackStore) receiveTracks() {
	for tracks := range s.ch {
		if err := s.handleTracks(tracks); err != nil {
			s.log.Error("could not handle updates", "err", err)
		}
	}
}

func (s *TrackStore) handleTracks(tracks []Track) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.tracks = tracks
	s.tracksSet = true
	return nil
}

func (s *TrackStore) processTracks() {
	for {
		track, ok := s.getNextTrack()
		if !ok {
			s.log.Debug("no tracks to convert")
			<-time.After(scanEvery)
			continue
		} else {
			s.log.Debug("converting a track", "track", track)
			if err := s.convert(track); err != nil {
				s.log.Error("conversion failed", "err", err)
				<-time.After(time.Second)
			}
		}
	}
}

func (s *TrackStore) getNextTrack() (Track, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	rand.Shuffle(len(s.tracks), func(i, j int) {
		s.tracks[i], s.tracks[j] = s.tracks[j], s.tracks[i]
	})

	for _, track := range s.tracks {
		p := s.filePath(track.Id)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			return track, true
		}
	}
	return Track{}, false
}

func (s *TrackStore) convert(track Track) error {
	outputPath := s.filePath(track.Id)
	tmpOutputPath := s.tmpFilePath(track.Id)

	if err := makeDirectory(outputPath); err != nil {
		return errors.Wrap(err, "could not create output directory")
	}

	args := []string{
		"-y",
		"-i",
		track.Path,
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
	s.log.Debug("converting", "command", cmd.String())
	if err := cmd.Run(); err != nil {
		s.log.Error("command error", "stderr", bufErr.String())
		return errors.Wrap(err, "ffmpeg execution failed")
	}

	if err := os.Rename(tmpOutputPath, outputPath); err != nil {
		return errors.Wrap(err, "move failed")
	}

	return nil
}

func (s *TrackStore) checkDuration(id string) (time.Duration, error) {
	filePath := s.filePath(id)

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
	s.log.Debug("checking duration", "command", cmd.String())
	output, err := cmd.Output()
	if err != nil {
		s.log.Error("command error", "stderr", bufErr.String())
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

func (s *TrackStore) filePath(id string) string {
	dir := path.Join(s.cacheDir, trackDirectory)
	file := fmt.Sprintf("%s.%s", id, trackExtension)
	return path.Join(dir, file)
}

func (s *TrackStore) tmpFilePath(id string) string {
	dir := path.Join(s.cacheDir, trackDirectory)
	file := fmt.Sprintf("_%s.%s", id, trackExtension)
	return path.Join(dir, file)
}
