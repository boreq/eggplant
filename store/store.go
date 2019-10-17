package store

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"sync"
	"time"

	"github.com/boreq/eggplant/logging"
	"github.com/pkg/errors"
)

var log = logging.New("store")

type Track struct {
	Path string
	Id   string
}

func New(cacheDir string, tracks []Track) (*Store, error) {
	store := &Store{
		tracks:   tracks,
		cacheDir: cacheDir,
	}
	go store.run()
	return store, nil
}

type Store struct {
	cacheDir string
	tracks   []Track
	mutex    sync.Mutex
}

func (s *Store) ServeFile(w http.ResponseWriter, r *http.Request, id string) {
	http.ServeFile(w, r, s.filePath(id))
}

func (s *Store) run() {
	for {
		track, ok := s.getNextTrackForConversion()
		if !ok {
			log.Debug("no tracks to convert")
			<-time.After(time.Second)
			continue
		}

		if err := s.convert(track); err != nil {
			log.Error("conversion failed", "err", err)
		}
	}
}

func (s *Store) getNextTrackForConversion() (Track, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, track := range s.tracks {
		p := s.filePath(track.Id)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			return track, true
		}
	}
	return Track{}, false
}

func (s *Store) convert(track Track) error {
	output := s.filePath(track.Id)
	if err := s.makeDirectory(output); err != nil {
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
		output,
	}
	cmd := exec.Command("ffmpeg", args...)
	bufErr := &bytes.Buffer{}
	cmd.Stderr = bufErr
	println(cmd.String())
	log.Debug("converting", "command", cmd.String())
	if err := cmd.Run(); err != nil {
		log.Error("command error", "stderr", bufErr.String())
		return errors.Wrap(err, "ffmpeg execution failed")
	}
	log.Debug("produced", "path", output)
	return nil
}

func (s *Store) filePath(id string) string {
	dir := path.Join(s.cacheDir, "tracks")
	file := fmt.Sprintf("%s.opus", id)
	return path.Join(dir, file)
}

func (s *Store) makeDirectory(p string) error {
	dir, _ := path.Split(p)
	return os.MkdirAll(dir, os.ModePerm)
}
