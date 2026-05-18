package domain

import (
	"github.com/boreq/errors"
)

type RootAlbum struct {
	thumbnail *Thumbnail
	albums    []ChildAlbum
	tracks    Tracks
}

func NewRootAlbum(thumbnail *Thumbnail, albums []ChildAlbum, tracks []Track) (RootAlbum, error) {
	return RootAlbum{
		thumbnail: thumbnail,
		albums:    albums,
		tracks:    NewTracks(tracks),
	}, nil
}

func (r RootAlbum) Thumbnail() *Thumbnail {
	return r.thumbnail
}

func (r RootAlbum) Albums() []ChildAlbum {
	return r.albums
}

func (r RootAlbum) Tracks() Tracks {
	return r.tracks
}

type Album struct {
	id        AlbumId
	title     AlbumTitle
	thumbnail *Thumbnail
	parents   []ParentAlbum
	albums    []ChildAlbum
	tracks    Tracks
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

func (a Album) Parents() []ParentAlbum {
	return a.parents
}

func (a Album) Albums() []ChildAlbum {
	return a.albums
}

func (a Album) Tracks() Tracks {
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
