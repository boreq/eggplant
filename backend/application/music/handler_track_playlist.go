package music

import (
	"context"

	"github.com/boreq/eggplant/domain"
	"github.com/boreq/eggplant/domain/library"
	"github.com/boreq/errors"
)

type TrackPlaylist struct {
	Id domain.TrackId
}

type TrackPlaylistHandler struct {
	libraryRepository LibraryRepository
	trackStore        TrackStore
}

func NewTrackPlaylistHandler(libraryRepository LibraryRepository, trackStore TrackStore) *TrackPlaylistHandler {
	return &TrackPlaylistHandler{
		libraryRepository: libraryRepository,
		trackStore:        trackStore,
	}
}

func (h *TrackPlaylistHandler) Execute(ctx context.Context, accessCtx library.AccessContext, cmd TrackPlaylist) (domain.ConvertedFile, error) {
	lib, err := h.libraryRepository.Get()
	if err != nil {
		return domain.ConvertedFile{}, errors.Wrap(err, "could not get the library")
	}

	track, err := lib.GetTrack(accessCtx, cmd.Id)
	if err != nil {
		return domain.ConvertedFile{}, errors.Wrap(err, "could not get the track")
	}

	cf, err := h.trackStore.GetPlaylist(ctx, track.FileId())
	if err != nil {
		return domain.ConvertedFile{}, errors.Wrap(err, "could not get the playlist")
	}

	return cf, nil
}
