package music

import "github.com/boreq/errors"

type TrackHandler struct {
	trackStore TrackStore
}

func NewTrackHandler(trackStore TrackStore) *TrackHandler {
	return &TrackHandler{
		trackStore: trackStore,
	}
}

func (h *TrackHandler) Execute(id string) (string, error) {
	p, err := h.trackStore.GetFilePath(id)
	if err != nil {
		return "", errors.Wrap(err, "could not get the track path")
	}
	return p, nil
}
