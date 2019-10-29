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

	"github.com/boreq/eggplant/logging"
	"github.com/boreq/eggplant/scanner"
	"github.com/boreq/eggplant/store"
	"github.com/pkg/errors"
)

const rootAlbumTitle = "Eggplant"

type AlbumId string

func (id AlbumId) String() string {
	return string(id)
}

type TrackId string

func (id TrackId) String() string {
	return string(id)
}

type FileId string

func (id FileId) String() string {
	return string(id)
}

type Thumbnail struct {
	FileId FileId `json:"fileId,omitempty"`
}

type Track struct {
	Id       TrackId `json:"id,omitempty"`
	FileId   FileId  `json:"fileId,omitempty"`
	Title    string  `json:"title,omitempty"`
	Duration float64 `json:"duration,omitempty"`
}

type Album struct {
	Id        AlbumId    `json:"id,omitempty"`
	Title     string     `json:"title,omitempty"`
	Thumbnail *Thumbnail `json:"thumbnail,omitempty"`
	Access    Access     `json:"access,omitempty"`

	Parents []Album `json:"parents,omitempty"`
	Albums  []Album `json:"albums,omitempty"`
	Tracks  []Track `json:"tracks,omitempty"`
}

type Access struct {
	NoPublic bool `json:"noPublic"`
}

type track struct {
	title  string
	path   string
	fileId FileId
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
	thumbnailId   FileId
	access        Access
	albums        map[AlbumId]*album
	tracks        map[TrackId]track
}

func newAlbum(title string) *album {
	return &album{
		title:  title,
		albums: make(map[AlbumId]*album),
		tracks: make(map[TrackId]track),
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
func (l *Library) Browse(ids []AlbumId) (Album, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	album, err := l.getAlbum(ids)
	if err != nil {
		return Album{}, errors.Wrap(err, "failed to get an album")
	}

	parents, err := l.getParents(ids)
	if err != nil {
		return Album{}, errors.Wrap(err, "failed to get parents")
	}

	listed := Album{
		Title:   album.title,
		Parents: parents,
		Access:  album.access,
	}

	if len(ids) > 0 {
		listed.Id = ids[len(ids)-1]
	}

	if album.thumbnailId != "" {
		listed.Thumbnail = &Thumbnail{
			FileId: album.thumbnailId,
		}
	}

	for id, album := range album.albums {
		d := Album{
			Id:     id,
			Title:  album.title,
			Access: album.access,
		}
		if album.thumbnailId != "" {
			d.Thumbnail = &Thumbnail{
				FileId: album.thumbnailId,
			}
		}
		listed.Albums = append(listed.Albums, d)
	}
	sort.Slice(listed.Albums, func(i, j int) bool { return listed.Albums[i].Title < listed.Albums[j].Title })

	for id, track := range album.tracks {
		t := Track{
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

func (l *Library) getParents(ids []AlbumId) ([]Album, error) {
	parents := make([]Album, 0)
	for i := 0; i < len(ids); i++ {
		parentIds := ids[:i+1]
		dir, err := l.getAlbum(parentIds)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get a parent album")
		}
		parent := Album{
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

func (l *Library) getAlbum(ids []AlbumId) (*album, error) {
	var current *album = l.root
	for _, id := range ids {
		child, ok := current.albums[id]
		if !ok {
			return nil, errors.Errorf("album '%s' not found", id)
		}
		current = child
	}
	return current, nil
}

func (l *Library) loadAccess(file string) (Access, error) {
	f, err := os.Open(file)
	if err != nil {
		return Access{}, errors.Wrap(err, "could not open the file")
	}
	defer f.Close()

	acc := Access{}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		switch scanner.Text() {
		case "no-public":
			acc.NoPublic = true
		default:
			return Access{}, fmt.Errorf("unrecognized line: %s", scanner.Text())
		}
	}

	if err := scanner.Err(); err != nil {
		return Access{}, errors.Wrap(err, "scanner error")
	}

	return acc, nil
}

func toTrack(title string, scannerTrack scanner.Track) (TrackId, track, error) {
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

func toAlbum(title string, scannerAlbum scanner.Album) (AlbumId, *album, error) {
	id, err := newAlbumId(title)
	if err != nil {
		return "", nil, errors.Wrap(err, "could not create an album id")
	}
	album := newAlbum(title)
	return id, album, nil
}

func newAlbumId(title string) (AlbumId, error) {
	h, err := shortHash(title)
	if err != nil {
		return "", errors.Wrap(err, "hashing failed")
	}
	return AlbumId(h), nil
}

func newTrackId(title string) (TrackId, error) {
	h, err := shortHash(title)
	if err != nil {
		return "", errors.Wrap(err, "hashing failed")
	}
	return TrackId(h), nil
}

func newFileId(path string) (FileId, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return "", errors.Wrap(err, "os stat failed")
	}
	s := fmt.Sprintf("%s-%d-%d", path, fileInfo.Size(), fileInfo.ModTime().Unix())
	h, err := longHash(s)
	if err != nil {
		return "", errors.Wrap(err, "hashing failed")
	}
	return FileId(h), nil
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
