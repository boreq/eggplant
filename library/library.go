package library

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"sort"
	"sync"

	"github.com/boreq/eggplant/loader"
	"github.com/boreq/eggplant/logging"
	"github.com/boreq/eggplant/store"
	"github.com/pkg/errors"
)

const rootAlbumTitle = "Eggplant"

type Id string

func (id Id) String() string {
	return string(id)
}

type Track struct {
	Id       string  `json:"id,omitempty"`
	Title    string  `json:"title,omitempty"`
	FileHash string  `json:"fileHash,omitempty"`
	Duration float64 `json:"duration,omitempty"`
}

type Album struct {
	Id    string `json:"id,omitempty"`
	Title string `json:"title,omitempty"`

	Parents []Album `json:"parents,omitempty"`
	Albums  []Album `json:"albums,omitempty"`
	Tracks  []Track `json:"tracks,omitempty"`
}

type track struct {
	title string
	path  string
}

func newTrack(title string, path string) track {
	t := track{
		title: title,
		path:  path,
	}
	return t
}

type album struct {
	title         string
	thumbnailPath string
	thumbnailId   Id
	albums        map[Id]*album
	tracks        map[Id]track
}

func newAlbum(title string) *album {
	return &album{
		title:  title,
		albums: make(map[Id]*album),
		tracks: make(map[Id]track),
	}
}

type Library struct {
	root           *album
	store          *store.Store
	thumbnailStore *store.ThumbnailStore
	mutex          sync.Mutex
	log            logging.Logger
}

func New(ch <-chan loader.Album, thumbnailStore *store.ThumbnailStore) (*Library, error) {
	l := &Library{
		log:            logging.New("library"),
		root:           newAlbum(rootAlbumTitle),
		thumbnailStore: thumbnailStore,
	}
	go l.receiveLoaderUpdates(ch)
	return l, nil

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
	defer l.mutex.Lock()

	// update the track list
	l.root = newAlbum(rootAlbumTitle)
	if err := l.mergeAlbum(l.root, album); err != nil {
		return errors.Wrap(err, "merge album failed")
	}

	// schedule track conversion

	// schedule thumbnail conversion
	var thumbnails []store.Thumbnail
	if err := l.getThumbnails(&thumbnails, l.root); err != nil {
		return errors.Wrap(err, "preparing a thumbnail list failed")
	}
	l.thumbnailStore.SetThumbnails(thumbnails)

	return nil
}

func (l *Library) mergeAlbum(target *album, album loader.Album) error {
	thumbnailId, err := longId(album.Thumbnail)
	if err != nil {
		return errors.Wrap(err, "could not create a thumbnail id")
	}
	target.thumbnailPath = album.Thumbnail
	target.thumbnailId = thumbnailId

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

func (l *Library) getThumbnails(thumbnails *[]store.Thumbnail, current *album) error {
	if current.thumbnailPath != "" {
		thumbnail := store.Thumbnail{
			Id:   current.thumbnailId.String(),
			Path: current.thumbnailPath,
		}
		*thumbnails = append(*thumbnails, thumbnail)
	}

	for _, child := range current.albums {
		l.getThumbnails(thumbnails, child)
	}

	return nil
}

func toTrack(title string, loaderTrack loader.Track) (Id, track, error) {
	id, err := shortId(loaderTrack.Path)
	if err != nil {
		return "", track{}, errors.Wrap(err, "could not create an id")
	}
	track := newTrack(title, loaderTrack.Path)
	return id, track, nil
}

func toAlbum(title string, loaderAlbum loader.Album) (Id, *album, error) {
	id, err := shortId(title)
	if err != nil {
		return "", nil, errors.Wrap(err, "could not create an id")
	}
	album := newAlbum(title)
	return id, album, nil
}

func (l *Library) getAlbum(ids []Id) (*album, error) {
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

func (l *Library) Browse(ids []Id) (Album, error) {
	listed := Album{}

	for i := 0; i < len(ids); i++ {
		parentIds := ids[:i+1]
		dir, err := l.getAlbum(parentIds)
		if err != nil {
			return Album{}, errors.Wrap(err, "failed to get directory")
		}
		parent := Album{
			Id:    parentIds[len(parentIds)-1].String(),
			Title: dir.title,
		}
		listed.Parents = append(listed.Parents, parent)
	}

	dir, err := l.getAlbum(ids)
	if err != nil {
		return Album{}, errors.Wrap(err, "failed to get directory")
	}

	listed.Title = dir.title

	for id, directory := range dir.albums {
		d := Album{
			Id:    id.String(),
			Title: directory.title,
		}
		listed.Albums = append(listed.Albums, d)
	}
	sort.Slice(listed.Albums, func(i, j int) bool { return listed.Albums[i].Title < listed.Albums[j].Title })

	for id, track := range dir.tracks {
		t := Track{
			Id:       id.String(),
			Title:    track.title,
			Duration: l.store.GetDuration(id.String()).Seconds(),
		}
		listed.Tracks = append(listed.Tracks, t)
	}
	sort.Slice(listed.Tracks, func(i, j int) bool { return listed.Tracks[i].Title < listed.Tracks[j].Title })

	return listed, nil
}

//func (l *Library) List() []store.Track {
//	m := make(map[string]store.Track)
//	l.list(m, l.root)
//	var tracks []store.Track
//	for _, track := range m {
//		tracks = append(tracks, track)
//	}
//	return tracks
//}

//func (l *Library) list(tracks map[string]store.Track, dir *directory) {
//	for _, track := range dir.tracks {
//		tracks[track.fileHash] = store.Track{
//			Path: track.path,
//			Id:   track.fileHash,
//		}
//	}
//	for _, subdirectory := range dir.directories {
//		l.list(tracks, subdirectory)
//	}
//}

func longId(s string) (Id, error) {
	sum, err := getHash(s)
	if err != nil {
		return "", errors.Wrap(err, "hashing failed")
	}
	return Id(hex.EncodeToString(sum)), nil
}

func shortId(s string) (Id, error) {
	sum, err := getHash(s)
	if err != nil {
		return "", errors.Wrap(err, "hashing failed")
	}
	return Id(hex.EncodeToString(sum)[:20]), nil
}

func getHash(s string) ([]byte, error) {
	buf := bytes.NewBuffer([]byte(s))
	hasher := sha256.New()
	if _, err := io.Copy(hasher, buf); err != nil {
		return nil, err
	}
	var sum []byte
	sum = hasher.Sum(sum)
	return sum, nil
}

//func getFileHash(p string) (string, error) {
//	f, err := os.Open(p)
//	if err != nil {
//		return "", err
//	}
//	defer f.Close()
//	hasher := sha256.New()
//	if _, err := io.Copy(hasher, f); err != nil {
//		return "", err
//	}
//	var sum []byte
//	sum = hasher.Sum(sum)
//	return hex.EncodeToString(sum), nil
//}

//func hashToId(sum []byte) Id {
//	return Id(hex.EncodeToString(sum)[:idLength])
//}
