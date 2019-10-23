package library

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"sort"
	"sync"

	"github.com/boreq/eggplant/loader"
	"github.com/boreq/eggplant/logging"
	"github.com/boreq/eggplant/store"
	"github.com/pkg/errors"
)

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
	Thumbnail *Thumbnail `json:"thumbnail,imitempty"`

	Parents []Album `json:"parents,omitempty"`
	Albums  []Album `json:"albums,omitempty"`
	Tracks  []Track `json:"tracks,omitempty"`
}

const rootAlbumTitle = "Eggplant"

type track struct {
	title  string
	path   string
	fileId FileId
}

func newTrack(title string, path string) (track, error) {
	fileId, err := newFileId(path)
	if err != nil {
		return track{}, errors.Wrap(err, "could not create file id")
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

type Library struct {
	root           *album
	trackStore     *store.TrackStore
	thumbnailStore *store.Store
	mutex          sync.Mutex
	log            logging.Logger
}

func New(ch <-chan loader.Album, thumbnailStore *store.Store, trackStore *store.TrackStore) (*Library, error) {
	l := &Library{
		log:            logging.New("library"),
		root:           newAlbum(rootAlbumTitle),
		thumbnailStore: thumbnailStore,
		trackStore:     trackStore,
	}
	go l.receiveLoaderUpdates(ch)
	return l, nil

}

func (l *Library) Browse(ids []AlbumId) (Album, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	album, err := l.getAlbum(ids)
	if err != nil {
		return Album{}, errors.Wrap(err, "failed to get directory")
	}

	parents, err := l.getParents(ids)
	if err != nil {
		return Album{}, errors.Wrap(err, "failed to get parents")
	}

	listed := Album{
		Title:   album.title,
		Parents: parents,
	}

	if album.thumbnailId != "" {
		listed.Thumbnail = &Thumbnail{
			FileId: album.thumbnailId,
		}
	}

	for id, album := range album.albums {
		d := Album{
			Id:    id,
			Title: album.title,
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

func (l *Library) receiveLoaderUpdates(ch <-chan loader.Album) {
	for album := range ch {
		if err := l.handleLoaderUpdate(album); err != nil {
			l.log.Error("could not handle a loader update", "err", err)
		}
	}
}

func (l *Library) handleLoaderUpdate(album loader.Album) error {
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

func (l *Library) mergeAlbum(target *album, album loader.Album) error {
	if album.Thumbnail != "" {
		thumbnailId, err := newFileId(album.Thumbnail)
		if err != nil {
			return errors.Wrap(err, "could not create a thumbnail id")
		}
		target.thumbnailPath = album.Thumbnail
		target.thumbnailId = thumbnailId
	}

	for title, loaderTrack := range album.Tracks {
		id, track, err := toTrack(title, loaderTrack)
		if err != nil {
			return errors.Wrap(err, "could not convert to a track")
		}
		target.tracks[id] = track
	}

	for title, loaderAlbum := range album.Albums {
		id, album, err := toAlbum(title, *loaderAlbum)
		if err != nil {
			return errors.Wrap(err, "could not convert to an album")
		}
		target.albums[id] = album
		l.mergeAlbum(album, *loaderAlbum)
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

func toTrack(title string, loaderTrack loader.Track) (TrackId, track, error) {
	id, err := newTrackId(title)
	if err != nil {
		return "", track{}, errors.Wrap(err, "could not create a track id")
	}
	t, err := newTrack(title, loaderTrack.Path)
	if err != nil {
		return "", track{}, errors.Wrap(err, "could not create a track")
	}
	return id, t, nil
}

func toAlbum(title string, loaderAlbum loader.Album) (AlbumId, *album, error) {
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
	s := fmt.Sprintf("%s-%s-%s", path, fileInfo.Size(), fileInfo.ModTime())
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
