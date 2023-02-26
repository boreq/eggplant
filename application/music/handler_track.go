package music

import (
	"context"
	"io"
	"time"

	"github.com/boreq/errors"
)

type ConvertedFile struct {
	// Name is just a filename used for mimetype detection. It is here just to
	// check its extension type basically.
	Name string

	// Modtime is used to figure out if the content has changed.
	Modtime time.Time

	// Content must be closed by the caller.
	Content io.ReadSeekCloser
}

type TrackHandler struct {
	trackStore TrackStore
}

func NewTrackHandler(trackStore TrackStore) *TrackHandler {
	return &TrackHandler{
		trackStore: trackStore,
	}
}

func (h *TrackHandler) Execute(ctx context.Context, id string) (ConvertedFile, error) {
	p, err := h.trackStore.GetConvertedFile(ctx, id)
	if err != nil {
		return ConvertedFile{}, errors.Wrap(err, "could not get the track")
	}
	return p, nil
}
