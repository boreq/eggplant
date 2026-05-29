package music

import (
	"context"

	"github.com/boreq/eggplant/domain/music"
	"github.com/boreq/eggplant/domain/music/library"
	"github.com/boreq/errors"
)

type StartStreaming struct {
	TrackId      music.TrackId
	SeekPosition *music.RequestedSeekPosition
}

type StartStreamingHandler struct {
	libraryRepository LibraryRepository
	trackConverter    TrackConverter
}

func NewStartStreamingHandler(libraryRepository LibraryRepository, trackConverter TrackConverter) *StartStreamingHandler {
	return &StartStreamingHandler{
		libraryRepository: libraryRepository,
		trackConverter:    trackConverter,
	}
}

func (h *StartStreamingHandler) Execute(ctx context.Context, accessCtx library.AccessContext, cmd StartStreaming) (music.StreamId, error) {
	lib, err := h.libraryRepository.Get()
	if err != nil {
		return music.StreamId{}, errors.Wrap(err, "could not get the library")
	}

	track, err := lib.GetTrack(accessCtx, cmd.TrackId)
	if err != nil {
		return music.StreamId{}, errors.Wrap(err, "could not get the track")
	}

	streamId, err := h.trackConverter.StartStream(ctx, track.FileId(), h.seekPosition(cmd))
	if err != nil {
		return music.StreamId{}, errors.Wrap(err, "could not start the stream")
	}

	return streamId, nil
}

func (h *StartStreamingHandler) seekPosition(cmd StartStreaming) *music.SeekPosition {
	if cmd.SeekPosition == nil {
		return nil
	}
	seekPosition := music.NewSeekPosition(*cmd.SeekPosition)
	return &seekPosition
}
