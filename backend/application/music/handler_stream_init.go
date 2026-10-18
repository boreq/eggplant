package music

import (
	"github.com/boreq/eggplant/domain/music"
	"github.com/boreq/eggplant/domain/music/library"
	"github.com/boreq/errors"
)

type StreamInit struct {
	trackId  music.TrackId
	streamId music.StreamId
}

func NewStreamInit(trackId music.TrackId, streamId music.StreamId) StreamInit {
	return StreamInit{
		trackId:  trackId,
		streamId: streamId,
	}
}

func (s StreamInit) TrackId() music.TrackId {
	return s.trackId
}

func (s StreamInit) StreamId() music.StreamId {
	return s.streamId
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

	track, err := lib.GetTrack(accessCtx, cmd.TrackId())
	if err != nil {
		return ConvertedFile{}, errors.Wrap(err, "could not get the track")
	}

	fileId, ok := track.FileId()
	if !ok {
		return ConvertedFile{}, errors.New("track has no file id")
	}

	cf, err := h.trackConverter.GetInit(fileId, cmd.StreamId())
	if err != nil {
		return ConvertedFile{}, errors.Wrap(err, "could not get the init segment")
	}
	return cf, nil
}
