package music

import (
	"context"
	"io"

	"github.com/boreq/eggplant/application/accessctx"
	"github.com/boreq/eggplant/domain/music"
	"github.com/boreq/eggplant/domain/remote"
	"github.com/boreq/errors"
)

type RemoteStreamInit struct {
	InstanceId remote.RemoteInstanceID
	TrackId    music.TrackId
	StreamId   music.StreamId
}

type RemoteStreamInitHandler struct {
	remoteLibrary RemoteLibrary
}

func NewRemoteStreamInitHandler(remoteLibrary RemoteLibrary) *RemoteStreamInitHandler {
	return &RemoteStreamInitHandler{remoteLibrary: remoteLibrary}
}

func (h *RemoteStreamInitHandler) Execute(ctx context.Context, accessCtx accessctx.AccessContext, cmd RemoteStreamInit) (io.ReadCloser, error) {
	if !accessCtx.Can(accessctx.PermissionSeeRemoteLibraryContent) {
		return nil, accessctx.ErrPermissionDenied
	}

	body, err := h.remoteLibrary.GetStreamInit(ctx, cmd.InstanceId, cmd.TrackId, cmd.StreamId)
	if err != nil {
		return nil, errors.Wrap(err, "could not get the remote stream init")
	}

	return body, nil
}
