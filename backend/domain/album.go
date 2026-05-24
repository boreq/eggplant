package domain

import (
	"github.com/boreq/errors"
)

type RootAlbum struct {
	thumbnail *Thumbnail
	albums    []PartialAlbum
	tracks    Tracks
}

func NewRootAlbum(thumbnail *Thumbnail, albums []PartialAlbum, tracks []Track) (RootAlbum, error) {
	return RootAlbum{
		thumbnail: thumbnail,
		albums:    albums,
		tracks:    NewTracks(tracks),
	}, nil
}

func (r RootAlbum) Thumbnail() *Thumbnail {
	return r.thumbnail
}

func (r RootAlbum) Albums() []PartialAlbum {
	return r.albums
}

func (r RootAlbum) Tracks() Tracks {
	return r.tracks
}

type Album struct {
	id        AlbumId
	title     AlbumTitle
	thumbnail *Thumbnail
	parents   []PartialAlbum
	albums    []PartialAlbum
	tracks    Tracks
}

func NewAlbum(id AlbumId, title AlbumTitle, thumbnail *Thumbnail, parents []PartialAlbum, albums []PartialAlbum, tracks []Track) (Album, error) {
	for _, p := range parents {
		if p.id == id {
			return Album{}, errors.New("parents must not contain the album itself")
		}
	}
	for _, c := range albums {
		if c.id == id {
			return Album{}, errors.New("child albums must not contain the album itself")
		}
	}
	if len(albums) == 0 && len(tracks) == 0 {
		return Album{}, errors.New("album must have at least one child album or track")
	}
	return Album{
		id:        id,
		title:     title,
		thumbnail: thumbnail,
		parents:   parents,
		albums:    albums,
		tracks:    NewTracks(tracks),
	}, nil
}

func (a Album) Id() AlbumId {
	return a.id
}

func (a Album) Title() AlbumTitle {
	return a.title
}

func (a Album) Thumbnail() *Thumbnail {
	return a.thumbnail
}

func (a Album) Parents() []PartialAlbum {
	return a.parents
}

func (a Album) Albums() []PartialAlbum {
	return a.albums
}

func (a Album) Tracks() Tracks {
	return a.tracks
}

type PartialAlbum struct {
	id        AlbumId
	title     AlbumTitle
	thumbnail *Thumbnail
}

func NewPartialAlbum(id AlbumId, title AlbumTitle, thumbnail *Thumbnail) PartialAlbum {
	return PartialAlbum{
		id:        id,
		title:     title,
		thumbnail: thumbnail,
	}
}

func (a PartialAlbum) Id() AlbumId {
	return a.id
}

func (a PartialAlbum) Title() AlbumTitle {
	return a.title
}

func (a PartialAlbum) Thumbnail() *Thumbnail {
	return a.thumbnail
}

type AlbumId struct {
	id idForHumans
}

func NewAlbumId(parents []AlbumId, title AlbumTitle) (AlbumId, error) {
	return AlbumId{id: newIdForHumans(parents, title)}, nil
}

func NewAlbumIdFromString(s string) (AlbumId, error) {
	id, err := newIdForHumansFromString(s)
	if err != nil {
		return AlbumId{}, errors.Wrap(err, "invalid album id")
	}
	return AlbumId{id: id}, nil
}

func (a AlbumId) String() string {
	return a.id.String()
}

type AlbumTitle struct {
	value string
}

func NewAlbumTitle(s string) (AlbumTitle, error) {
	if s == "" {
		return AlbumTitle{}, errors.New("album title must not be empty")
	}
	return AlbumTitle{value: s}, nil
}

func (t AlbumTitle) String() string {
	return t.value
}
