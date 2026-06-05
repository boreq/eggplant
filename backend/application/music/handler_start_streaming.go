package music

import (
	"context"

	"github.com/boreq/eggplant/domain/music"
	"github.com/boreq/eggplant/domain/music/library"
	"github.com/boreq/eggplant/internal/logging"
	"github.com/boreq/errors"
)

var startStreamingLog = logging.New("StartStreamingHandler")

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
	startStreamingLog.Debug("Execute: entered", "trackId", cmd.TrackId.String())

	lib, err := h.libraryRepository.Get()
	if err != nil {
		return music.StreamId{}, errors.Wrap(err, "could not get the library")
	}
	startStreamingLog.Debug("Execute: got library")

	track, err := lib.GetTrack(accessCtx, cmd.TrackId)
	if err != nil {
		return music.StreamId{}, errors.Wrap(err, "could not get the track")
	}
	startStreamingLog.Debug("Execute: got track", "fileId", track.FileId().String())

	startStreamingLog.Debug("Execute: calling trackConverter.StartStream")
	streamId, err := h.trackConverter.StartStream(ctx, track.FileId(), h.seekPosition(cmd))
	startStreamingLog.Debug("Execute: trackConverter.StartStream returned", "err", err)
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
