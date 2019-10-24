package loader

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/boreq/eggplant/logging"
	"github.com/pkg/errors"
	"github.com/radovskyb/watcher"
)

type Track struct {
	Path string
}

func newTrack(path string) Track {
	return Track{
		Path: path,
	}
}

type Album struct {
	Thumbnail string
	Albums    map[string]*Album
	Tracks    map[string]Track
}

func newAlbum() *Album {
	return &Album{
		Albums: make(map[string]*Album),
		Tracks: make(map[string]Track),
	}
}

type Loader struct {
	directory string
	log       logging.Logger
}

func New(directory string) (*Loader, error) {
	l := &Loader{
		directory: directory,
		log:       logging.New("loader"),
	}
	return l, nil
}

func (l *Loader) Start() (<-chan Album, error) {
	// fail early since the initial load carries the highest failure
	// possiblity
	album, err := l.load()
	if err != nil {
		return nil, errors.Wrap(err, "initial load failed")
	}

	w := watcher.New()
	w.SetMaxEvents(1)

	if err := w.AddRecursive(l.directory); err != nil {
		return nil, errors.Wrap(err, "could not add a watcher")
	}

	go func() {
		if err := w.Start(time.Second * 10); err != nil {
			l.log.Error("watcher start returned an error", "err", err)
		}
	}()

	ch := make(chan Album)
	go func() {
		ch <- album

		for {
			select {
			case <-w.Event:
				l.log.Debug("reloading")
				album, err := l.load()
				if err != nil {
					l.log.Error("load error", "err", err)
					continue
				}
				ch <- album
			case err := <-w.Error:
				l.log.Error("watcher error", "err", err)
			case <-w.Closed:
				return
			}
			<-time.After(10 * time.Second) // in case something goes crazy
		}
	}()
	return ch, nil
}

func (l *Loader) load() (Album, error) {
	root := *newAlbum()
	if err := filepath.Walk(l.directory, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			if l.isThumbnail(path) {
				if err := l.addThumbnail(&root, path); err != nil {
					return errors.Wrap(err, "could not add a thumbnail")
				}
			} else {
				if err := l.addTrack(&root, path); err != nil {
					return errors.Wrap(err, "could not add a track")
				}
			}
		}
		return nil
	}); err != nil {
		return Album{}, errors.Wrap(err, "walk failed")
	}
	return root, nil
}

func (l *Loader) addTrack(root *Album, file string) error {
	album, err := l.findAlbum(root, file)
	if err != nil {
		return errors.Wrap(err, "could not find an album")
	}

	title := filenameWithoutExtension(file)
	album.Tracks[title] = newTrack(file)
	return nil
}

func (l *Loader) addThumbnail(root *Album, file string) error {
	album, err := l.findAlbum(root, file)
	if err != nil {
		return errors.Wrap(err, "could not find an album")
	}

	album.Thumbnail = file
	return nil
}

func (l *Loader) isThumbnail(path string) bool {
	filename := filenameWithoutExtension(path)
	return filename == "thumbnail"
}

func (l *Loader) findAlbum(root *Album, file string) (*Album, error) {
	relativePath, err := filepath.Rel(l.directory, file)
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
