package music

import (
	"context"

	"github.com/boreq/eggplant/domain"
	"github.com/boreq/eggplant/domain/library"
	"github.com/boreq/errors"
)

type TrackInit struct {
	Id domain.TrackId
}

type TrackInitHandler struct {
	libraryRepository LibraryRepository
	trackStore        TrackStore
}

func NewTrackInitHandler(libraryRepository LibraryRepository, trackStore TrackStore) *TrackInitHandler {
	return &TrackInitHandler{
		libraryRepository: libraryRepository,
		trackStore:        trackStore,
	}
}

func (h *TrackInitHandler) Execute(ctx context.Context, accessCtx library.AccessContext, cmd TrackInit) (domain.ConvertedFile, error) {
	lib, err := h.libraryRepository.Get()
	if err != nil {
		return domain.ConvertedFile{}, errors.Wrap(err, "could not get the library")
	}

	track, err := lib.GetTrack(accessCtx, cmd.Id)
	if err != nil {
		return domain.ConvertedFile{}, errors.Wrap(err, "could not get the track")
	}

	cf, err := h.trackStore.GetInit(ctx, track.FileId())
	if err != nil {
		return domain.ConvertedFile{}, errors.Wrap(err, "could not get the init segment")
	}

	return cf, nil
}
