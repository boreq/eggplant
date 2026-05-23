package queries

type UserRepository interface {
	Count() (int, error)
}

type TrackStore interface {
	GetStats() (TrackStats, error)
}

type ThumbnailStore interface {
	GetStats() (ThumbnailStats, error)
}

type Stats struct {
	Users      int            `json:"users"`
	Thumbnails ThumbnailStats `json:"thumbnails"`
	Tracks     TrackStats     `json:"tracks"`
}

type TrackStats struct {
	NumberOfTracks int64 `json:"numberOfTracks"`
	SizeOfTracks   int64 `json:"sizeOfTracks"`

	NumberOfStreams       int64 `json:"numberOfStreams"`
	SizeOfConvertedTracks int64 `json:"sizeOfConvertedTracks"`
}

type ThumbnailStats struct {
	AllItems       int64 `json:"allItems"`
	ConvertedItems int64 `json:"convertedItems"`
	OriginalSize   int64 `json:"originalSize"`
	ConvertedSize  int64 `json:"convertedSize"`
}

type TransactionProvider interface {
	Read(handler TransactionHandler) error
}

type TransactionHandler func(repositories *TransactableRepositories) error

type TransactableRepositories struct {
	Users UserRepository
}
