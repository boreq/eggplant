package queries

import "github.com/boreq/errors"

type StatsHandler struct {
	trackStore          TrackStore
	thumbnailStore      ThumbnailStore
	transactionProvider TransactionProvider
}

func NewStatsHandler(
	trackStore TrackStore,
	thumbnailStore ThumbnailStore,
	transactionProvider TransactionProvider,
) *StatsHandler {
	return &StatsHandler{
		trackStore:          trackStore,
		thumbnailStore:      thumbnailStore,
		transactionProvider: transactionProvider,
	}
}

func (h *StatsHandler) Execute() (Stats, error) {
	var users int
	if err := h.transactionProvider.Read(func(r *TransactableRepositories) error {
		n, err := r.Users.Count()
		if err != nil {
			return errors.Wrap(err, "count failed")
		}
		users = n
		return nil
	}); err != nil {
		return Stats{}, errors.Wrap(err, "transaction failed")
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
