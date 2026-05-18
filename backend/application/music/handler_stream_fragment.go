package music

import (
	"github.com/boreq/eggplant/domain"
	"github.com/boreq/eggplant/domain/library"
	"github.com/boreq/errors"
)

type StreamFragment struct {
	TrackId    domain.TrackId
	StreamId   domain.StreamId
	FragmentId domain.FragmentId
}

type StreamFragmentHandler struct {
	libraryRepository LibraryRepository
	trackConverter    TrackConverter
}

func NewStreamFragmentHandler(libraryRepository LibraryRepository, trackConverter TrackConverter) *StreamFragmentHandler {
	return &StreamFragmentHandler{
		libraryRepository: libraryRepository,
		trackConverter:    trackConverter,
	}
}

func (h *StreamFragmentHandler) Execute(accessCtx library.AccessContext, cmd StreamFragment) (ConvertedFile, error) {
	lib, err := h.libraryRepository.Get()
	if err != nil {
		return ConvertedFile{}, errors.Wrap(err, "could not get the library")
	}

	track, err := lib.GetTrack(accessCtx, cmd.TrackId)
	if err != nil {
		return ConvertedFile{}, errors.Wrap(err, "could not get the track")
	}

	cf, err := h.trackConverter.GetFragment(track.FileId(), cmd.StreamId, cmd.FragmentId)
	if err != nil {
		return ConvertedFile{}, errors.Wrap(err, "could not get the fragment")
	}
	return cf, nil
}
