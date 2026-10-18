package music

import (
	"github.com/boreq/eggplant/domain/music"
	"github.com/boreq/eggplant/domain/music/library"
	"github.com/boreq/errors"
)

type StreamPlaylist struct {
	trackId  music.TrackId
	streamId music.StreamId
}

func NewStreamPlaylist(trackId music.TrackId, streamId music.StreamId) StreamPlaylist {
	return StreamPlaylist{
		trackId:  trackId,
		streamId: streamId,
	}
}

func (s StreamPlaylist) TrackId() music.TrackId {
	return s.trackId
}

func (s StreamPlaylist) StreamId() music.StreamId {
	return s.streamId
}

type StreamPlaylistHandler struct {
	libraryRepository LibraryRepository
	trackConverter    TrackConverter
}

func NewStreamPlaylistHandler(libraryRepository LibraryRepository, trackConverter TrackConverter) *StreamPlaylistHandler {
	return &StreamPlaylistHandler{
		libraryRepository: libraryRepository,
		trackConverter:    trackConverter,
	}
}

func (h *StreamPlaylistHandler) Execute(accessCtx library.AccessContext, cmd StreamPlaylist) (Playlist, error) {
	lib, err := h.libraryRepository.Get()
	if err != nil {
		return Playlist{}, errors.Wrap(err, "could not get the library")
	}

	track, err := lib.GetTrack(accessCtx, cmd.TrackId())
	if err != nil {
		return Playlist{}, errors.Wrap(err, "could not get the track")
	}

	fileId, ok := track.FileId()
	if !ok {
		return Playlist{}, errors.New("track has no file id")
	}

	playlist, err := h.trackConverter.GetPlaylist(fileId, cmd.StreamId())
	if err != nil {
		return Playlist{}, errors.Wrap(err, "could not get the playlist")
	}

	return playlist, nil
}
