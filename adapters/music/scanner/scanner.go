// Package scanner is responsible for generating a tree-like structure of
// albums and tracks based on the contents of a directory.
package scanner

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/boreq/eggplant/logging"
	"github.com/boreq/errors"
	"github.com/radovskyb/watcher"
)

// Track represents an audio track.
type Track struct {
	Path string
}

func newTrack(path string) Track {
	return Track{
		Path: path,
	}
}

// Album is a collection of songs and albums.
type Album struct {
	// Thumbnail is a path of an album cover. If the thumbnail is not
	// available then this field is set to an empty string.
	Thumbnail string

	// AccessFile is a path of an access file. If the access file is not
	// available then this field is set to an empty string.
	AccessFile string

	// Albums uses album titles as its keys.
	Albums map[string]*Album

	// Tracks uses track titles as its keys.
	Tracks map[string]Track
}

func newAlbum() *Album {
	return &Album{
		Albums: make(map[string]*Album),
		Tracks: make(map[string]Track),
	}
}

// Scanner watches a hard drive directory containing audio files and produces
// updates whenever its contents change.
type Scanner struct {
	directory string
	log       logging.Logger
}

// New creates a new scanner which will watch the specified directory when
// started.
func New(directory string) (*Scanner, error) {
	l := &Scanner{
		directory: directory,
		log:       logging.New("scanner"),
	}
	return l, nil
}

// Start starts the watcher and returns a channel on which the updates are
// sent whenever available. At least one update will be sent on the channel
// immidiately after calling this method.
func (s *Scanner) Start() (<-chan Album, error) {
	// fail early since the initial load carries the highest failure
	// possiblity
	album, err := s.load()
	if err != nil {
		return nil, errors.Wrap(err, "initial load failed")
	}

	w := watcher.New()
	w.SetMaxEvents(1)

	if err := w.AddRecursive(s.directory); err != nil {
		return nil, errors.Wrap(err, "could not add a recursive watcher")
	}

	go func() {
		if err := w.Start(time.Second * 10); err != nil {
			s.log.Error("error starting the watcher", "err", err)
		}
	}()

	ch := make(chan Album)
	go func() {
		defer close(ch)
		ch <- album

		for {
			select {
			case _, ok := <-w.Event:
				if !ok {
					return
				}
				album, err := s.load()
				if err != nil {
					s.log.Error("load error", "err", err)
					continue
				}
				if len(album.Tracks) < 1 {
					// empty album
					continue
				}
				ch <- album
			case err := <-w.Error:
				s.log.Error("watcher error", "err", err)
			case <-w.Closed:
				return
			}
		}
	}()
	return ch, nil
}

func (s *Scanner) load() (Album, error) {
	root := *newAlbum()
	albumEmpty := true
	if err := filepath.Walk(s.directory, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			if s.isThumbnail(path) {
				if err := s.addThumbnail(&root, path); err != nil {
					return errors.Wrap(err, "could not add a thumbnail")
				}
				return nil
			}

			if s.isAccessFile(path) {
				if err := s.addAccessFile(&root, path); err != nil {
					return errors.Wrap(err, "could not add an access file")
				}
				return nil
			}

			if s.isTrack(path) {
				if err := s.addTrack(&root, path); err != nil {
					return errors.Wrap(err, "could not add a track")
				}
				albumEmpty = false
			}
		}
		return nil
	}); err != nil {
		return Album{}, errors.Wrap(err, "walk failed")
	}
	if albumEmpty {
		return *newAlbum(), nil
	}
	return root, nil
}

func (s *Scanner) addTrack(root *Album, file string) error {
	album, err := s.findAlbum(root, file)
	if err != nil {
		return errors.Wrap(err, "could not find an album")
	}

	title := filenameWithoutExtension(file)
	album.Tracks[title] = newTrack(file)
	return nil
}

func (s *Scanner) addThumbnail(root *Album, file string) error {
	album, err := s.findAlbum(root, file)
	if err != nil {
		return errors.Wrap(err, "could not find an album")
	}

	album.Thumbnail = file
	return nil
}

func (s *Scanner) addAccessFile(root *Album, file string) error {
	album, err := s.findAlbum(root, file)
	if err != nil {
		return errors.Wrap(err, "could not find an album")
	}

	album.AccessFile = file
	return nil
}

func (s *Scanner) isAccessFile(path string) bool {
	_, filename := filepath.Split(path)
	return filename == "eggplant.access"
}

func (s *Scanner) isThumbnail(path string) bool {
	filename := filenameWithoutExtension(path)
	return filename == "thumbnail"
}

func (s *Scanner) isTrack(path string) bool {
	ext := strings.ToLower(filepath.Ext(path)[1:])

	return ext == "flac" || ext == "mp3" || ext == "ogg" || ext == "aac" || ext == "wav" || ext == "wma" || ext == "aiff"
}

func (s *Scanner) findAlbum(root *Album, file string) (*Album, error) {
	relativePath, err := filepath.Rel(s.directory, file)
	if err != nil {
		return nil, errors.Wrap(err, "could not get a relative filepath")
	}

	dir, _ := filepath.Split(relativePath)
	if dir == "" {
		return root, nil
	}

	names := strings.Split(strings.Trim(dir, string(os.PathSeparator)), string(os.PathSeparator))

	var album *Album = root
	for _, name := range names {
		child, ok := album.Albums[name]
		if !ok {
			child = newAlbum()
			album.Albums[name] = child
		}
		album = child
	}
	return album, nil
}

func filenameWithoutExtension(file string) string {
	_, filename := filepath.Split(file)
	if index := strings.LastIndex(filename, "."); index >= 0 {
		return filename[:index]
	}
	return filename
}
