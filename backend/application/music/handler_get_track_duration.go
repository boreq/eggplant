package music

import (
	"context"

	"github.com/boreq/eggplant/domain/music"
	"github.com/boreq/eggplant/domain/music/library"
	"github.com/boreq/errors"
)

type GetTrackDuration struct {
	trackId music.TrackId
}

func NewGetTrackDuration(trackId music.TrackId) GetTrackDuration {
	return GetTrackDuration{
		trackId: trackId,
	}
}

func (g GetTrackDuration) TrackId() music.TrackId {
	return g.trackId
}

type GetTrackDurationHandler struct {
	libraryRepository LibraryRepository
	durations         TrackDurationStore
}

func NewGetTrackDurationHandler(libraryRepository LibraryRepository, durations TrackDurationStore) *GetTrackDurationHandler {
	return &GetTrackDurationHandler{
		libraryRepository: libraryRepository,
		durations:         durations,
	}
}

func (h *GetTrackDurationHandler) Execute(ctx context.Context, accessCtx library.AccessContext, cmd GetTrackDuration) (music.TrackDuration, error) {
	lib, err := h.libraryRepository.Get()
	if err != nil {
		return music.TrackDuration{}, errors.Wrap(err, "could not get the library")
	}

	track, err := lib.GetTrack(accessCtx, cmd.TrackId())
	if err != nil {
		return music.TrackDuration{}, errors.Wrap(err, "could not get the track")
	}

	fileId, ok := track.FileId()
	if !ok {
		return music.TrackDuration{}, errors.New("track has no file id")
	}

	duration, err := h.durations.GetDuration(ctx, fileId)
	if err != nil {
		return music.TrackDuration{}, errors.Wrap(err, "could not get the duration")
	}

	return duration, nil
}
