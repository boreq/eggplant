package music

type ThumbnailStore interface {
	GetFilePath(id string) (string, error)
}

type TrackStore interface {
	GetFilePath(id string) (string, error)
}

type Library interface {
	Browse(ids []AlbumId) (Album, error)
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
