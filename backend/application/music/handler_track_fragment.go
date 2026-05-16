package music

import (
	"context"

	"github.com/boreq/eggplant/domain"
	"github.com/boreq/eggplant/domain/library"
	"github.com/boreq/errors"
)

type TrackFragment struct {
	Id         domain.TrackId
	FragmentId domain.TrackFragmentId
}

type TrackFragmentHandler struct {
	libraryRepository LibraryRepository
	trackStore        TrackStore
}

func NewTrackFragmentHandler(libraryRepository LibraryRepository, trackStore TrackStore) *TrackFragmentHandler {
	return &TrackFragmentHandler{
		libraryRepository: libraryRepository,
		trackStore:        trackStore,
	}
}

func (h *TrackFragmentHandler) Execute(ctx context.Context, accessCtx library.AccessContext, cmd TrackFragment) (domain.ConvertedFile, error) {
	lib, err := h.libraryRepository.Get()
	if err != nil {
		return domain.ConvertedFile{}, errors.Wrap(err, "could not get the library")
	}

	track, err := lib.GetTrack(accessCtx, cmd.Id)
	if err != nil {
		return domain.ConvertedFile{}, errors.Wrap(err, "could not get the track")
	}

	cf, err := h.trackStore.GetFragment(ctx, track.FileId(), cmd.FragmentId)
	if err != nil {
		return domain.ConvertedFile{}, errors.Wrap(err, "could not get the fragment")
	}

	return cf, nil
}
