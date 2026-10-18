package music

import (
	"github.com/boreq/eggplant/domain/music"
	"github.com/boreq/eggplant/domain/music/library"
	"github.com/boreq/errors"
)

type KeepAliveStream struct {
	trackId  music.TrackId
	streamId music.StreamId
}

func NewKeepAliveStream(trackId music.TrackId, streamId music.StreamId) KeepAliveStream {
	return KeepAliveStream{
		trackId:  trackId,
		streamId: streamId,
	}
}

func (k KeepAliveStream) TrackId() music.TrackId {
	return k.trackId
}

func (k KeepAliveStream) StreamId() music.StreamId {
	return k.streamId
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

	track, err := lib.GetTrack(accessCtx, cmd.TrackId())
	if err != nil {
		return errors.Wrap(err, "could not get the track")
	}

	fileId, ok := track.FileId()
	if !ok {
		return errors.New("track has no file id")
	}

	return h.trackConverter.KeepAliveStream(fileId, cmd.StreamId())
}
