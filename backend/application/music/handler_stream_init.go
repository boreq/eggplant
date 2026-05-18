package music

import (
	"github.com/boreq/eggplant/domain"
	"github.com/boreq/eggplant/domain/library"
	"github.com/boreq/errors"
)

type StreamInit struct {
	TrackId  domain.TrackId
	StreamId domain.StreamId
}

type StreamInitHandler struct {
	libraryRepository LibraryRepository
	trackConverter    TrackConverter
}

func NewStreamInitHandler(libraryRepository LibraryRepository, trackConverter TrackConverter) *StreamInitHandler {
	return &StreamInitHandler{
		libraryRepository: libraryRepository,
		trackConverter:    trackConverter,
	}
}

func (h *StreamInitHandler) Execute(accessCtx library.AccessContext, cmd StreamInit) (ConvertedFile, error) {
	lib, err := h.libraryRepository.Get()
	if err != nil {
		return ConvertedFile{}, errors.Wrap(err, "could not get the library")
	}

	track, err := lib.GetTrack(accessCtx, cmd.TrackId)
	if err != nil {
		return ConvertedFile{}, errors.Wrap(err, "could not get the track")
	}

	cf, err := h.trackConverter.GetInit(track.FileId(), cmd.StreamId)
	if err != nil {
		return ConvertedFile{}, errors.Wrap(err, "could not get the init segment")
	}
	return cf, nil
}
