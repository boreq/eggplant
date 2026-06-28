package music

import (
	"context"

	"github.com/boreq/eggplant/application/accessctx"
	"github.com/boreq/eggplant/domain/music"
	"github.com/boreq/eggplant/domain/remote"
	"github.com/boreq/errors"
)

type RemoteKeepAliveStream struct {
	InstanceId remote.RemoteInstanceID
	TrackId    music.TrackId
	StreamId   music.StreamId
}

type RemoteKeepAliveStreamHandler struct {
	remoteLibrary RemoteLibrary
}

func NewRemoteKeepAliveStreamHandler(remoteLibrary RemoteLibrary) *RemoteKeepAliveStreamHandler {
	return &RemoteKeepAliveStreamHandler{remoteLibrary: remoteLibrary}
}

func (h *RemoteKeepAliveStreamHandler) Execute(ctx context.Context, accessCtx accessctx.AccessContext, cmd RemoteKeepAliveStream) error {
	if !accessCtx.Can(accessctx.PermissionSeeRemoteLibraryContent) {
		return accessctx.ErrPermissionDenied
	}

	if err := h.remoteLibrary.KeepAliveStream(ctx, cmd.InstanceId, cmd.TrackId, cmd.StreamId); err != nil {
		return errors.Wrap(err, "could not keep alive the remote stream")
	}

	return nil
}
