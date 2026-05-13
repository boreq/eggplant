package domain

import (
	"encoding/hex"
	"errors"
)

type Album struct {
	id        AlbumId
	title     AlbumTitle
	thumbnail *Thumbnail
	access    Access

	// parents lists the parents of this album starting from the one furthest
	// away. If non-empty, the last entry is this album itself.
	parents []AlbumParent
	albums  []Album
	tracks  []Track
}

func NewAlbum(id AlbumId, title AlbumTitle, thumbnail *Thumbnail, access Access, parents []AlbumParent, albums []Album, tracks []Track) (Album, error) {
	if len(parents) > 0 && parents[len(parents)-1].Id() != id {
		return Album{}, errors.New("the last parent must be this album")
	}
	return Album{
		id:        id,
		title:     title,
		thumbnail: thumbnail,
		access:    access,
		parents:   parents,
		albums:    albums,
		tracks:    tracks,
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

func (a Album) Access() Access {
	return a.access
}

func (a Album) Parents() []AlbumParent {
	return a.parents
}

func (a Album) Albums() []Album {
	return a.albums
}

func (a Album) Tracks() []Track {
	return a.tracks
}

type AlbumParent struct {
	id    AlbumId
	title AlbumTitle
}

func NewAlbumParent(id AlbumId, title AlbumTitle) AlbumParent {
	return AlbumParent{
		id:    id,
		title: title,
	}
}

func (p AlbumParent) Id() AlbumId {
	return p.id
}

func (p AlbumParent) Title() AlbumTitle {
	return p.title
}

type Access struct {
	public bool
}

func NewAccess(public bool) Access {
	return Access{public: public}
}

func (a Access) Public() bool {
	return a.public
}

type AlbumId struct {
	value string
}

func NewAlbumId(s string) (AlbumId, error) {
	if s == "" {
		return AlbumId{}, errors.New("album id must not be empty")
	}
	if _, err := hex.DecodeString(s); err != nil {
		return AlbumId{}, errors.New("album id must be a hex string")
	}
	return AlbumId{value: s}, nil
}

func (id AlbumId) String() string {
	return id.value
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
