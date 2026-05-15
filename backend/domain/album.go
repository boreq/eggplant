package domain

import (
	"encoding/hex"
	"errors"
)

type RootAlbum struct {
	thumbnail *Thumbnail
	albums    []ChildAlbum
	tracks    []Track
}

func NewRootAlbum(thumbnail *Thumbnail, albums []ChildAlbum, tracks []Track) (RootAlbum, error) {
	return RootAlbum{
		thumbnail: thumbnail,
		albums:    albums,
		tracks:    tracks,
	}, nil
}

func (r RootAlbum) Thumbnail() *Thumbnail {
	return r.thumbnail
}

func (r RootAlbum) Albums() []ChildAlbum {
	return r.albums
}

func (r RootAlbum) Tracks() []Track {
	return r.tracks
}

type Album struct {
	id        AlbumId
	title     AlbumTitle
	thumbnail *Thumbnail
	parents   []ParentAlbum
	albums    []ChildAlbum
	tracks    []Track
}

func NewAlbum(id AlbumId, title AlbumTitle, thumbnail *Thumbnail, parents []ParentAlbum, albums []ChildAlbum, tracks []Track) (Album, error) {
	if len(parents) == 0 {
		return Album{}, errors.New("album must have at least one parent")
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

func (a Album) Parents() []ParentAlbum {
	return a.parents
}

func (a Album) Albums() []ChildAlbum {
	return a.albums
}

func (a Album) Tracks() []Track {
	return a.tracks
}

type ChildAlbum struct {
	id        AlbumId
	title     AlbumTitle
	thumbnail *Thumbnail
}

func NewChildAlbum(id AlbumId, title AlbumTitle, thumbnail *Thumbnail) ChildAlbum {
	return ChildAlbum{
		id:        id,
		title:     title,
		thumbnail: thumbnail,
	}
}

func (a ChildAlbum) Id() AlbumId {
	return a.id
}

func (a ChildAlbum) Title() AlbumTitle {
	return a.title
}

func (a ChildAlbum) Thumbnail() *Thumbnail {
	return a.thumbnail
}

type ParentAlbum struct {
	id    AlbumId
	title AlbumTitle
}

func NewParentAlbum(id AlbumId, title AlbumTitle) ParentAlbum {
	return ParentAlbum{
		id:    id,
		title: title,
	}
}

func (p ParentAlbum) Id() AlbumId {
	return p.id
}

func (p ParentAlbum) Title() AlbumTitle {
	return p.title
}

type AlbumId struct {
	value string
}

func NewAlbumIdFromString(s string) (AlbumId, error) {
	if s == "" {
		return AlbumId{}, errors.New("album id must not be empty")
	}
	if _, err := hex.DecodeString(s); err != nil {
		return AlbumId{}, errors.New("album id must be a hex string")
	}
	return AlbumId{value: s}, nil
}

func NewAlbumId(parents []AlbumId, title AlbumTitle) (AlbumId, error) {
	return NewAlbumIdFromString(shortHash(parentsAsString(parents) + title.value))
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
