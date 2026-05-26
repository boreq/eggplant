package music

import (
	"github.com/boreq/eggplant/domain/music"
	"github.com/boreq/eggplant/domain/music/library"
	"github.com/boreq/errors"
)

type StreamPlaylist struct {
	TrackId  music.TrackId
	StreamId music.StreamId
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

	track, err := lib.GetTrack(accessCtx, cmd.TrackId)
	if err != nil {
		return Playlist{}, errors.Wrap(err, "could not get the track")
	}

	playlist, err := h.trackConverter.GetPlaylist(track.FileId(), cmd.StreamId)
	if err != nil {
		return Playlist{}, errors.Wrap(err, "could not get the playlist")
	}

	return playlist, nil
}
