package music

import (
	"context"

	"github.com/boreq/eggplant/application/accessctx"
	"github.com/boreq/eggplant/domain/music"
	"github.com/boreq/eggplant/domain/remote"
	"github.com/boreq/errors"
)

type RemoteStartStreaming struct {
	instanceId   remote.RemoteInstanceID
	trackId      music.TrackId
	seekPosition *music.RequestedSeekPosition
}

func NewRemoteStartStreaming(instanceId remote.RemoteInstanceID, trackId music.TrackId, seekPosition *music.RequestedSeekPosition) RemoteStartStreaming {
	return RemoteStartStreaming{
		instanceId:   instanceId,
		trackId:      trackId,
		seekPosition: seekPosition,
	}
}

func (r RemoteStartStreaming) InstanceId() remote.RemoteInstanceID {
	return r.instanceId
}

func (r RemoteStartStreaming) TrackId() music.TrackId {
	return r.trackId
}

func (r RemoteStartStreaming) SeekPosition() *music.RequestedSeekPosition {
	return r.seekPosition
}

type RemoteStartStreamingHandler struct {
	remoteLibrary RemoteLibrary
}

func NewRemoteStartStreamingHandler(remoteLibrary RemoteLibrary) *RemoteStartStreamingHandler {
	return &RemoteStartStreamingHandler{remoteLibrary: remoteLibrary}
}

func (h *RemoteStartStreamingHandler) Execute(ctx context.Context, accessCtx accessctx.AccessContext, cmd RemoteStartStreaming) (music.StreamId, error) {
	if !accessCtx.Can(accessctx.PermissionSeeRemoteLibraryContent) {
		return music.StreamId{}, accessctx.ErrPermissionDenied
	}

	streamId, err := h.remoteLibrary.StartTrackStream(ctx, cmd.InstanceId(), cmd.TrackId(), cmd.SeekPosition())
	if err != nil {
		return music.StreamId{}, errors.Wrap(err, "could not start the remote stream")
	}

	return streamId, nil
}
