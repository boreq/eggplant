package music

import (
	"context"

	"github.com/boreq/eggplant/domain"
	"github.com/boreq/eggplant/domain/library"
	"github.com/boreq/errors"
)

type TrackHandler struct {
	libraryRepository LibraryRepository
	trackStore        TrackStore
}

func NewTrackHandler(libraryRepository LibraryRepository, trackStore TrackStore) *TrackHandler {
	return &TrackHandler{
		libraryRepository: libraryRepository,
		trackStore:        trackStore,
	}
}

func (h *TrackHandler) Execute(ctx context.Context, accessCtx library.AccessContext, id domain.TrackId) (domain.ConvertedFile, error) {
	lib, err := h.libraryRepository.Get()
	if err != nil {
		return domain.ConvertedFile{}, errors.Wrap(err, "could not get the library")
	}

	track, err := lib.GetTrack(id, accessCtx)
	if err != nil {
		return domain.ConvertedFile{}, errors.Wrap(err, "could not get the track")
	}

	cf, err := h.trackStore.GetConvertedFile(ctx, track.FileId())
	if err != nil {
		return domain.ConvertedFile{}, errors.Wrap(err, "could not get the converted file")
	}

	return cf, nil
}
