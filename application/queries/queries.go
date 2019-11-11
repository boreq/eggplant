package queries

type UserRepository interface {
	Count() (int, error)
}

type TrackStore interface {
	GetStats() (StoreStats, error)
}

type ThumbnailStore interface {
	GetStats() (StoreStats, error)
}

type Stats struct {
	Users      int        `json:"users"`
	Thumbnails StoreStats `json:"thumbnails"`
	Tracks     StoreStats `json:"tracks"`
}

type StoreStats struct {
	AllItems       int   `json:"allItems"`
	ConvertedItems int   `json:"convertedItems"`
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
