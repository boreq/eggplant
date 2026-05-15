package music

import (
	"context"

	"github.com/boreq/eggplant/domain"
	"github.com/boreq/errors"
)

type TrackHandler struct {
	trackStore TrackStore
}

func NewTrackHandler(trackStore TrackStore) *TrackHandler {
	return &TrackHandler{
		trackStore: trackStore,
	}
}

func (h *TrackHandler) Execute(ctx context.Context, id domain.TrackId) (domain.ConvertedFile, error) {
	p, err := h.trackStore.GetConvertedFile(ctx, id)
	if err != nil {
		return domain.ConvertedFile{}, errors.Wrap(err, "could not get the track")
	}
	return p, nil
}
