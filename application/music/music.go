package music

type ThumbnailStore interface {
	GetFilePath(id string) (string, error)
}

type TrackStore interface {
	GetFilePath(id string) (string, error)
}
