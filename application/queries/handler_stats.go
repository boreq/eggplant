package queries

import "github.com/boreq/errors"

type StatsHandler struct {
	userRepository UserRepository
	trackStore     TrackStore
	thumbnailStore ThumbnailStore
}

func NewStatsHandler(
	userRepository UserRepository,
	trackStore TrackStore,
	thumbnailStore ThumbnailStore,
) *StatsHandler {
	return &StatsHandler{
		userRepository: userRepository,
		trackStore:     trackStore,
		thumbnailStore: thumbnailStore,
	}
}

func (h *StatsHandler) Execute() (Stats, error) {
	users, err := h.userRepository.Count()
	if err != nil {
		return Stats{}, errors.Wrap(err, "could not count the users")
	}

	tracks, err := h.trackStore.GetStats()
	if err != nil {
		return Stats{}, errors.Wrap(err, "could not get the track stas")
	}

	thumbnails, err := h.thumbnailStore.GetStats()
	if err != nil {
		return Stats{}, errors.Wrap(err, "could not get the track stas")
	}

	stats := Stats{
		Users:      users,
		Thumbnails: thumbnails,
		Tracks:     tracks,
	}

	return stats, nil
}
