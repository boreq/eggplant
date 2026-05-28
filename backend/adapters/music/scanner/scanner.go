// Package scanner is responsible for generating a tree-like structure of
// albums and tracks based on the contents of a directory.
package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/boreq/eggplant/adapters/music/scanner/symwalk"
	"github.com/boreq/eggplant/domain/music"
	"github.com/boreq/eggplant/domain/music/scanner"
	"github.com/boreq/eggplant/logging"
	"github.com/boreq/errors"
)

type ThumbnailStem struct {
	value string
}

func NewThumbnailStem(s string) (ThumbnailStem, error) {
	if s == "" {
		return ThumbnailStem{}, errors.New("thumbnail stem must not be empty")
	}
	return ThumbnailStem{value: s}, nil
}

func (t ThumbnailStem) String() string {
	return t.value
}

type Config struct {
	trackExtensions     []music.FileExtension
	thumbnailStems      []ThumbnailStem
	thumbnailExtensions []music.FileExtension
}

func NewConfig(trackExtensions []music.FileExtension, thumbnailStems []ThumbnailStem, thumbnailExtensions []music.FileExtension) (Config, error) {
	if len(trackExtensions) == 0 {
		return Config{}, errors.New("missing track extensions")
	}

	if len(thumbnailStems) == 0 {
		return Config{}, errors.New("missing thumbnail stems")
	}

	if len(thumbnailExtensions) == 0 {
		return Config{}, errors.New("missing thumbnail extensions")
	}

	return Config{
		trackExtensions:     trackExtensions,
		thumbnailStems:      thumbnailStems,
		thumbnailExtensions: thumbnailExtensions,
	}, nil
}

type Scanner struct {
	directory string
	config    Config
	logger    logging.Logger
}

func New(directory string, config Config) (*Scanner, error) {
	l := &Scanner{
		directory: directory,
		config:    config,
		logger:    logging.New("scanner"),
	}
	return l, nil
}

func (s *Scanner) Scan() (scanner.FoundRootAlbum, error) {
	visited := make(map[string]struct{})

	root := newAlbum()
	if err := symwalk.Walk(s.directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrap(err, "received an error")
		}

		realPath, err := filepath.EvalSymlinks(path)
		if err != nil {
			return errors.Wrap(err, "could not eval a symlink")
		}

		_, ok := visited[realPath]
		if ok {
			return fmt.Errorf("loop detected: '%s' visited multiple times", realPath)
		}

		visited[realPath] = struct{}{}

		if info.Mode()&os.ModeDir != 0 { // skip directories
			return nil
		}

		if info.Mode()&os.ModeSymlink != 0 { // skip symlinks
			return nil
		}

		filePath, err := music.NewFilePath(path)
		if err != nil {
			return errors.Wrap(err, "could not create file path")
		}

		if s.isThumbnail(filePath) {
			if err := s.setThumbnailFile(root, filePath); err != nil {
				return errors.Wrap(err, "could not add a thumbnail")
			}
			return nil
		}

		if s.isAccessFile(path) {
			if err := s.setAccessFile(root, filePath); err != nil {
				return errors.Wrap(err, "could not add an access file")
			}
			return nil
		}

		if s.isTrack(filePath) {
			if err := s.addTrack(root, filePath, info.Size()); err != nil {
				return errors.Wrap(err, "could not add a track")
			}
			return nil
		}

		return nil
	}); err != nil {
		return scanner.FoundRootAlbum{}, errors.Wrap(err, "walk failed")
	}

	removeEmptyAlbums(root)

	result, err := toFoundRootAlbum(root)
	if err != nil {
		return scanner.FoundRootAlbum{}, errors.Wrap(err, "could not build directory scan result")
	}
	return result, nil
}

func (s *Scanner) addTrack(root *album, file music.FilePath, size int64) error {
	a, err := s.findAlbum(root, file)
	if err != nil {
		return errors.Wrap(err, "could not find an album")
	}

	base := filepath.Base(file.String())
	stripped := strings.TrimSuffix(base, filepath.Ext(base))

	title, err := music.NewTrackTitle(stripped)
	if err != nil {
		return errors.Wrap(err, "could not create track title")
	}

	if existing, exists := a.tracks[title]; exists {
		kept, discarded := existing, track{path: file, size: size}
		if size > existing.size {
			kept, discarded = discarded, kept
		}
		s.logger.Warn("duplicate track title, keeping larger file",
			"title", title,
			"kept", kept,
			"discarded", discarded,
		)
		a.tracks[title] = kept
		return nil
	}
	a.tracks[title] = track{path: file, size: size}
	return nil
}

func (s *Scanner) setThumbnailFile(root *album, file music.FilePath) error {
	a, err := s.findAlbum(root, file)
	if err != nil {
		return errors.Wrap(err, "could not find an album")
	}
	if a.thumbnailFile != nil {
		return fmt.Errorf("thumbnail file already set (%s) but found a new one (%s)", *a.thumbnailFile, file)
	}
	a.thumbnailFile = &file
	return nil
}

func (s *Scanner) setAccessFile(root *album, file music.FilePath) error {
	a, err := s.findAlbum(root, file)
	if err != nil {
		return errors.Wrap(err, "could not find an album")
	}
	if a.accessFile != nil {
		return fmt.Errorf("access file already set (%s) but found a new one (%s)", *a.accessFile, file)
	}
	a.accessFile = &file
	return nil
}

func (s *Scanner) isAccessFile(path string) bool {
	_, filename := filepath.Split(path)
	return filename == "eggplant.access"
}

func (s *Scanner) isThumbnail(path music.FilePath) bool {
	for _, thumbnailStem := range s.config.thumbnailStems {
		for _, thumbnailExt := range s.config.thumbnailExtensions {
			name := thumbnailStem.String() + thumbnailExt.String()
			if strings.EqualFold(filepath.Base(path.String()), name) {
				return true
			}
		}
	}
	return false
}

func (s *Scanner) isTrack(path music.FilePath) bool {
	for _, trackExt := range s.config.trackExtensions {
		if path.HasExtension(trackExt) {
			return true
		}
	}
	return false
}

func (s *Scanner) findAlbum(root *album, file music.FilePath) (*album, error) {
	relativePath, err := filepath.Rel(s.directory, file.String())
	if err != nil {
		return nil, errors.Wrap(err, "could not get a relative filepath")
	}

	dir, _ := filepath.Split(relativePath)
	if dir == "" {
		return root, nil
	}

	names := strings.Split(strings.Trim(dir, string(os.PathSeparator)), string(os.PathSeparator))

	current := root
	for _, name := range names {
		title, err := music.NewAlbumTitle(name)
		if err != nil {
			return nil, errors.Wrapf(err, "could not create album title for '%s'", name)
		}
		child, ok := current.albums[title]
		if !ok {
			child = newAlbum()
			current.albums[title] = child
		}
		current = child
	}
	return current, nil
}

type album struct {
	thumbnailFile *music.FilePath
	accessFile    *music.FilePath
	albums        map[music.AlbumTitle]*album
	tracks        map[music.TrackTitle]track
}

func newAlbum() *album {
	return &album{
		albums: map[music.AlbumTitle]*album{},
		tracks: map[music.TrackTitle]track{},
	}
}

type track struct {
	path music.FilePath
	size int64
}

func removeEmptyAlbums(root *album) {
	for title, a := range root.albums {
		removeEmptyAlbums(a)

		if len(a.albums) == 0 && len(a.tracks) == 0 {
			delete(root.albums, title)
		}
	}
}

func toFoundRootAlbum(root *album) (scanner.FoundRootAlbum, error) {
	albums, err := toFoundAlbums(root.albums)
	if err != nil {
		return scanner.FoundRootAlbum{}, err
	}
	return scanner.NewFoundRootAlbum(root.thumbnailFile, root.accessFile, albums, toFoundTracks(root.tracks)), nil
}

func toFoundAlbums(src map[music.AlbumTitle]*album) (map[music.AlbumTitle]scanner.FoundAlbum, error) {
	out := make(map[music.AlbumTitle]scanner.FoundAlbum, len(src))
	for title, a := range src {
		subAlbums, err := toFoundAlbums(a.albums)
		if err != nil {
			return nil, err
		}

		found, err := scanner.NewFoundAlbum(a.thumbnailFile, a.accessFile, subAlbums, toFoundTracks(a.tracks))
		if err != nil {
			return nil, errors.Wrapf(err, "could not create found album '%s'", title)
		}
		out[title] = found
	}
	return out, nil
}

func toFoundTracks(src map[music.TrackTitle]track) map[music.TrackTitle]scanner.FoundTrack {
	out := make(map[music.TrackTitle]scanner.FoundTrack, len(src))
	for title, t := range src {
		out[title] = scanner.NewFoundTrack(t.path)
	}
	return out
}
