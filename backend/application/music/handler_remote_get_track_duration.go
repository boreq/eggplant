package music

import (
	"context"

	"github.com/boreq/eggplant/application/accessctx"
	"github.com/boreq/eggplant/domain/music"
	"github.com/boreq/eggplant/domain/remote"
	"github.com/boreq/errors"
)

type RemoteGetTrackDuration struct {
	instanceId remote.RemoteInstanceID
	trackId    music.TrackId
}

func NewRemoteGetTrackDuration(instanceId remote.RemoteInstanceID, trackId music.TrackId) RemoteGetTrackDuration {
	return RemoteGetTrackDuration{
		instanceId: instanceId,
		trackId:    trackId,
	}
}

func (r RemoteGetTrackDuration) InstanceId() remote.RemoteInstanceID {
	return r.instanceId
}

func (r RemoteGetTrackDuration) TrackId() music.TrackId {
	return r.trackId
}

type RemoteGetTrackDurationHandler struct {
	remoteLibrary RemoteLibrary
}

func NewRemoteGetTrackDurationHandler(remoteLibrary RemoteLibrary) *RemoteGetTrackDurationHandler {
	return &RemoteGetTrackDurationHandler{remoteLibrary: remoteLibrary}
}

func (h *RemoteGetTrackDurationHandler) Execute(ctx context.Context, accessCtx accessctx.AccessContext, cmd RemoteGetTrackDuration) (music.TrackDuration, error) {
	if !accessCtx.Can(accessctx.PermissionSeeRemoteLibraryContent) {
		return music.TrackDuration{}, accessctx.ErrPermissionDenied
	}

	duration, err := h.remoteLibrary.GetTrackDuration(ctx, cmd.InstanceId(), cmd.TrackId())
	if err != nil {
		return music.TrackDuration{}, errors.Wrap(err, "could not get the remote track duration")
	}

	return duration, nil
}
