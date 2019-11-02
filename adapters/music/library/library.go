// Package library is responsible for orchestrating actions related to
// providing a navigable representation of the audio library.
package library

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"sort"
	"sync"

	"github.com/boreq/eggplant/adapters/music/scanner"
	"github.com/boreq/eggplant/adapters/music/store"
	"github.com/boreq/eggplant/application/music"
	"github.com/boreq/eggplant/logging"
	"github.com/boreq/errors"
)

const rootAlbumTitle = "Eggplant"

type track struct {
	title  string
	path   string
	fileId music.FileId
}

func newTrack(title string, path string) (track, error) {
	fileId, err := newFileId(path)
	if err != nil {
		return track{}, errors.Wrap(err, "could not create a file id")
	}
	t := track{
		title:  title,
		path:   path,
		fileId: fileId,
	}
	return t, nil
}

type album struct {
	title         string
	thumbnailPath string
	thumbnailId   music.FileId
	access        music.Access
	albums        map[music.AlbumId]*album
	tracks        map[music.TrackId]track
}

func newAlbum(title string) *album {
	return &album{
		title:  title,
		albums: make(map[music.AlbumId]*album),
		tracks: make(map[music.TrackId]track),
	}
}

// Library receives scanner updates, dispatches them to appropriate stores and
// builds a navigable representation of the music collection.
type Library struct {
	root           *album
	trackStore     *store.TrackStore
	thumbnailStore *store.Store
	mutex          sync.Mutex
	log            logging.Logger
}

// New creates a library which receives updates from the specified channel.
func New(ch <-chan scanner.Album, thumbnailStore *store.Store, trackStore *store.TrackStore) (*Library, error) {
	l := &Library{
		log:            logging.New("library"),
		root:           newAlbum(rootAlbumTitle),
		thumbnailStore: thumbnailStore,
		trackStore:     trackStore,
	}
	go l.receiveUpdates(ch)
	return l, nil

}

// Browse lists the specified album. Provide a zero-length slice to list the
// root album.
func (l *Library) Browse(ids []music.AlbumId) (music.Album, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	album, err := l.getAlbum(ids)
	if err != nil {
		return music.Album{}, errors.Wrap(err, "failed to get an album")
	}

	parents, err := l.getParents(ids)
	if err != nil {
		return music.Album{}, errors.Wrap(err, "failed to get parents")
	}

	listed := music.Album{
		Title:   album.title,
		Parents: parents,
		Access:  album.access,
	}

	if len(ids) > 0 {
		listed.Id = ids[len(ids)-1]
	}

	if album.thumbnailId != "" {
		listed.Thumbnail = &music.Thumbnail{
			FileId: album.thumbnailId,
		}
	}

	for id, album := range album.albums {
		d := music.Album{
			Id:     id,
			Title:  album.title,
			Access: album.access,
		}
		if album.thumbnailId != "" {
			d.Thumbnail = &music.Thumbnail{
				FileId: album.thumbnailId,
			}
		}
		listed.Albums = append(listed.Albums, d)
	}
	sort.Slice(listed.Albums, func(i, j int) bool { return listed.Albums[i].Title < listed.Albums[j].Title })

	for id, track := range album.tracks {
		t := music.Track{
			Id:       id,
			FileId:   track.fileId,
			Title:    track.title,
			Duration: l.trackStore.GetDuration(track.fileId.String()).Seconds(),
		}
		listed.Tracks = append(listed.Tracks, t)
	}
	sort.Slice(listed.Tracks, func(i, j int) bool { return listed.Tracks[i].Title < listed.Tracks[j].Title })

	return listed, nil
}

func (l *Library) getParents(ids []music.AlbumId) ([]music.Album, error) {
	parents := make([]music.Album, 0)
	for i := 0; i < len(ids); i++ {
		parentIds := ids[:i+1]
		dir, err := l.getAlbum(parentIds)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get a parent album")
		}
		parent := music.Album{
			Id:    parentIds[len(parentIds)-1],
			Title: dir.title,
		}
		parents = append(parents, parent)

	}
	return parents, nil
}

func (l *Library) receiveUpdates(ch <-chan scanner.Album) {
	for album := range ch {
		if err := l.handleUpdate(album); err != nil {
			l.log.Error("could not handle a scanner update", "err", err)
		}
	}
}

func (l *Library) handleUpdate(album scanner.Album) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// update the track list
	l.root = newAlbum(rootAlbumTitle)
	if err := l.mergeAlbum(l.root, album); err != nil {
		return errors.Wrap(err, "merge album failed")
	}

	// schedule track conversion
	var tracks []store.Item
	if err := l.getTracks(&tracks, l.root); err != nil {
		return errors.Wrap(err, "preparing tracks failed")
	}
	l.trackStore.SetItems(tracks)

	// schedule thumbnail conversion
	var thumbnails []store.Item
	if err := l.getThumbnails(&thumbnails, l.root); err != nil {
		return errors.Wrap(err, "preparing thumbnails failed")
	}
	l.thumbnailStore.SetItems(thumbnails)

	return nil
}

func (l *Library) mergeAlbum(target *album, album scanner.Album) error {
	if album.Thumbnail != "" {
		thumbnailId, err := newFileId(album.Thumbnail)
		if err != nil {
			return errors.Wrap(err, "could not create a thumbnail id")
		}
		target.thumbnailPath = album.Thumbnail
		target.thumbnailId = thumbnailId
	}

	if album.AccessFile != "" {
		acc, err := l.loadAccess(album.AccessFile)
		if err != nil {
			return errors.Wrap(err, "could not load the access file")
		}
		target.access = acc
	}

	for title, scannerTrack := range album.Tracks {
		id, track, err := toTrack(title, scannerTrack)
		if err != nil {
			return errors.Wrap(err, "could not convert to a track")
		}
		target.tracks[id] = track
	}

	for title, scannerAlbum := range album.Albums {
		id, album, err := toAlbum(title, *scannerAlbum)
		if err != nil {
			return errors.Wrap(err, "could not convert to an album")
		}
		target.albums[id] = album
		if err := l.mergeAlbum(album, *scannerAlbum); err != nil {
			return err
		}
	}

	return nil
}

func (l *Library) getThumbnails(thumbnails *[]store.Item, current *album) error {
	if current.thumbnailPath != "" {
		thumbnail := store.Item{
			Id:   current.thumbnailId.String(),
			Path: current.thumbnailPath,
		}
		*thumbnails = append(*thumbnails, thumbnail)
	}

	for _, child := range current.albums {
		if err := l.getThumbnails(thumbnails, child); err != nil {
			return err
		}
	}

	return nil
}

func (l *Library) getTracks(tracks *[]store.Item, current *album) error {
	for _, track := range current.tracks {
		track := store.Item{
			Id:   track.fileId.String(),
			Path: track.path,
		}
		*tracks = append(*tracks, track)
	}

	for _, child := range current.albums {
		if err := l.getTracks(tracks, child); err != nil {
			return err
		}
	}

	return nil
}

func (l *Library) getAlbum(ids []music.AlbumId) (*album, error) {
	var current *album = l.root
	for _, id := range ids {
		child, ok := current.albums[id]
		if !ok {
			return nil, fmt.Errorf("album '%s' not found", id)
		}
		current = child
	}
	return current, nil
}

func (l *Library) loadAccess(file string) (music.Access, error) {
	f, err := os.Open(file)
	if err != nil {
		return music.Access{}, errors.Wrap(err, "could not open the file")
	}
	defer f.Close()

	acc := music.Access{}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		switch scanner.Text() {
		case "no-public":
			acc.NoPublic = true
		default:
			return music.Access{}, fmt.Errorf("unrecognized line: %s", scanner.Text())
		}
	}

	if err := scanner.Err(); err != nil {
		return music.Access{}, errors.Wrap(err, "scanner error")
	}

	return acc, nil
}

func toTrack(title string, scannerTrack scanner.Track) (music.TrackId, track, error) {
	id, err := newTrackId(title)
	if err != nil {
		return "", track{}, errors.Wrap(err, "could not create a track id")
	}
	t, err := newTrack(title, scannerTrack.Path)
	if err != nil {
		return "", track{}, errors.Wrap(err, "could not create a track")
	}
	return id, t, nil
}

func toAlbum(title string, scannerAlbum scanner.Album) (music.AlbumId, *album, error) {
	id, err := newAlbumId(title)
	if err != nil {
		return "", nil, errors.Wrap(err, "could not create an album id")
	}
	album := newAlbum(title)
	return id, album, nil
}

func newAlbumId(title string) (music.AlbumId, error) {
	h, err := shortHash(title)
	if err != nil {
		return "", errors.Wrap(err, "hashing failed")
	}
	return music.AlbumId(h), nil
}

func newTrackId(title string) (music.TrackId, error) {
	h, err := shortHash(title)
	if err != nil {
		return "", errors.Wrap(err, "hashing failed")
	}
	return music.TrackId(h), nil
}

func newFileId(path string) (music.FileId, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return "", errors.Wrap(err, "os stat failed")
	}
	s := fmt.Sprintf("%s-%d-%d", path, fileInfo.Size(), fileInfo.ModTime().Unix())
	h, err := longHash(s)
	if err != nil {
		return "", errors.Wrap(err, "hashing failed")
	}
	return music.FileId(h), nil
}

func shortHash(s string) (string, error) {
	sum, err := longHash(s)
	if err != nil {
		return "", errors.Wrap(err, "hashing failed")
	}
	return sum[:20], nil
}

func longHash(s string) (string, error) {
	buf := bytes.NewBuffer([]byte(s))
	hasher := sha256.New()
	if _, err := io.Copy(hasher, buf); err != nil {
		return "", err
	}
	var sum []byte
	sum = hasher.Sum(sum)
	return hex.EncodeToString(sum), nil
}
