package music

import (
	"context"
	"io"

	"github.com/boreq/eggplant/application/accessctx"
	"github.com/boreq/eggplant/domain/music"
	"github.com/boreq/eggplant/domain/remote"
	"github.com/boreq/errors"
)

type RemoteStreamFragment struct {
	instanceId remote.RemoteInstanceID
	trackId    music.TrackId
	streamId   music.StreamId
	fragmentId music.FragmentId
}

func NewRemoteStreamFragment(instanceId remote.RemoteInstanceID, trackId music.TrackId, streamId music.StreamId, fragmentId music.FragmentId) RemoteStreamFragment {
	return RemoteStreamFragment{
		instanceId: instanceId,
		trackId:    trackId,
		streamId:   streamId,
		fragmentId: fragmentId,
	}
}

func (r RemoteStreamFragment) InstanceId() remote.RemoteInstanceID {
	return r.instanceId
}

func (r RemoteStreamFragment) TrackId() music.TrackId {
	return r.trackId
}

func (r RemoteStreamFragment) StreamId() music.StreamId {
	return r.streamId
}

func (r RemoteStreamFragment) FragmentId() music.FragmentId {
	return r.fragmentId
}

type RemoteStreamFragmentHandler struct {
	remoteLibrary RemoteLibrary
}

func NewRemoteStreamFragmentHandler(remoteLibrary RemoteLibrary) *RemoteStreamFragmentHandler {
	return &RemoteStreamFragmentHandler{remoteLibrary: remoteLibrary}
}

func (h *RemoteStreamFragmentHandler) Execute(ctx context.Context, accessCtx accessctx.AccessContext, cmd RemoteStreamFragment) (io.ReadCloser, error) {
	if !accessCtx.Can(accessctx.PermissionSeeRemoteLibraryContent) {
		return nil, accessctx.ErrPermissionDenied
	}

	body, err := h.remoteLibrary.GetStreamFragment(ctx, cmd.InstanceId(), cmd.TrackId(), cmd.StreamId(), cmd.FragmentId())
	if err != nil {
		return nil, errors.Wrap(err, "could not get the remote stream fragment")
	}

	return body, nil
}
