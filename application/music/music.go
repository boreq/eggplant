package music

import "errors"

var ErrForbidden = errors.New("forbidden")
var ErrNotFound = errors.New("not found")

type ThumbnailStore interface {
	GetFilePath(id string) (string, error)
}

type TrackStore interface {
	GetFilePath(id string) (string, error)
}

type SearchResult struct {
	Albums []BasicAlbum
	Tracks []SearchResultTrack
}

type BasicAlbum struct {
	Path      []AlbumId
	Title     string
	Thumbnail *Thumbnail
}

type SearchResultTrack struct {
	Track Track
	Album BasicAlbum
}

type Library interface {
	Browse(ids []AlbumId, publicOnly bool) (Album, error)
	Search(query string, publicOnly bool) (SearchResult, error)
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

	// Parents list the parents of this album starting from the one
	// furthest away from this album. The list of parent albums includes
	// this particular album. Only fields Id and Title are filled in.
	Parents []Album `json:"parents,omitempty"`
	Albums  []Album `json:"albums,omitempty"`
	Tracks  []Track `json:"tracks,omitempty"`
}

type Access struct {
	Public bool `json:"public"`
}

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
