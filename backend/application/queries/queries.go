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
	users      int
	thumbnails ThumbnailStats
	tracks     TrackStats
}

func NewStats(users int, thumbnails ThumbnailStats, tracks TrackStats) Stats {
	return Stats{
		users:      users,
		thumbnails: thumbnails,
		tracks:     tracks,
	}
}

func (s Stats) Users() int {
	return s.users
}

func (s Stats) Thumbnails() ThumbnailStats {
	return s.thumbnails
}

func (s Stats) Tracks() TrackStats {
	return s.tracks
}

type TrackStats struct {
	numberOfTracks int64
	sizeOfTracks   int64

	numberOfStreams int64
	streamCacheSize int64
}

func NewTrackStats(numberOfTracks, sizeOfTracks, numberOfStreams, streamCacheSize int64) TrackStats {
	return TrackStats{
		numberOfTracks:  numberOfTracks,
		sizeOfTracks:    sizeOfTracks,
		numberOfStreams: numberOfStreams,
		streamCacheSize: streamCacheSize,
	}
}

func (s TrackStats) NumberOfTracks() int64 {
	return s.numberOfTracks
}

func (s TrackStats) SizeOfTracks() int64 {
	return s.sizeOfTracks
}

func (s TrackStats) NumberOfStreams() int64 {
	return s.numberOfStreams
}

func (s TrackStats) StreamCacheSize() int64 {
	return s.streamCacheSize
}

type ThumbnailStats struct {
	allItems       int64
	convertedItems int64
	originalSize   int64
	convertedSize  int64
}

func NewThumbnailStats(allItems, convertedItems, originalSize, convertedSize int64) ThumbnailStats {
	return ThumbnailStats{
		allItems:       allItems,
		convertedItems: convertedItems,
		originalSize:   originalSize,
		convertedSize:  convertedSize,
	}
}

func (s ThumbnailStats) AllItems() int64 {
	return s.allItems
}

func (s ThumbnailStats) ConvertedItems() int64 {
	return s.convertedItems
}

func (s ThumbnailStats) OriginalSize() int64 {
	return s.originalSize
}

func (s ThumbnailStats) ConvertedSize() int64 {
	return s.convertedSize
}

type TransactionProvider interface {
	Read(handler TransactionHandler) error
}

type TransactionHandler func(repositories *TransactableRepositories) error

type TransactableRepositories struct {
	users UserRepository
}

func NewTransactableRepositories(
	users UserRepository,
) *TransactableRepositories {
	return &TransactableRepositories{
		users: users,
	}
}

func (r TransactableRepositories) Users() UserRepository {
	return r.users
}
