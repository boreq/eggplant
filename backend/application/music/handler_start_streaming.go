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

	seekPosition, err := h.seekPosition(cmd, track)
	if err != nil {
		return music.StreamId{}, errors.Wrap(err, "could not resolve seek position")
	}

	streamId, err := h.trackConverter.StartStream(ctx, track.FileId(), seekPosition)
	if err != nil {
		return music.StreamId{}, errors.Wrap(err, "could not start the stream")
	}

	return streamId, nil
}

func (h *StartStreamingHandler) seekPosition(cmd StartStreaming, track music.Track) (*music.SeekPosition, error) {
	if cmd.SeekPosition == nil {
		return nil, nil
	}
	seekPosition, err := music.NewSeekPosition(*cmd.SeekPosition, track)
	if err != nil {
		return nil, errors.Wrap(err, "could not validate seek position against the track")
	}
	return &seekPosition, nil
}
