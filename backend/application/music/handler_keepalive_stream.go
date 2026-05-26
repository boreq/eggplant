package music

import (
	"github.com/boreq/eggplant/domain"
	"github.com/boreq/eggplant/domain/library"
	"github.com/boreq/errors"
)

type KeepAliveStream struct {
	TrackId  domain.TrackId
	StreamId domain.StreamId
}

type KeepAliveStreamHandler struct {
	libraryRepository LibraryRepository
	trackConverter    TrackConverter
}

func NewKeepAliveStreamHandler(libraryRepository LibraryRepository, trackConverter TrackConverter) *KeepAliveStreamHandler {
	return &KeepAliveStreamHandler{
		libraryRepository: libraryRepository,
		trackConverter:    trackConverter,
	}
}

func (h *KeepAliveStreamHandler) Execute(accessCtx library.AccessContext, cmd KeepAliveStream) error {
	lib, err := h.libraryRepository.Get()
	if err != nil {
		return errors.Wrap(err, "could not get the library")
	}

	track, err := lib.GetTrack(accessCtx, cmd.TrackId)
	if err != nil {
		return errors.Wrap(err, "could not get the track")
	}

	return h.trackConverter.KeepAliveStream(track.FileId(), cmd.StreamId)
}
