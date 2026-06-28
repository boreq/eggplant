package music

import (
	"context"
	"io"

	"github.com/boreq/eggplant/application/accessctx"
	"github.com/boreq/eggplant/domain/music"
	"github.com/boreq/eggplant/domain/remote"
	"github.com/boreq/errors"
)

type RemoteStreamPlaylist struct {
	InstanceId remote.RemoteInstanceID
	TrackId    music.TrackId
	StreamId   music.StreamId
}

type RemoteStreamPlaylistHandler struct {
	remoteLibrary RemoteLibrary
}

func NewRemoteStreamPlaylistHandler(remoteLibrary RemoteLibrary) *RemoteStreamPlaylistHandler {
	return &RemoteStreamPlaylistHandler{remoteLibrary: remoteLibrary}
}

func (h *RemoteStreamPlaylistHandler) Execute(ctx context.Context, accessCtx accessctx.AccessContext, cmd RemoteStreamPlaylist) (io.ReadCloser, error) {
	if !accessCtx.Can(accessctx.PermissionSeeRemoteLibraryContent) {
		return nil, accessctx.ErrPermissionDenied
	}

	body, err := h.remoteLibrary.GetStreamPlaylist(ctx, cmd.InstanceId, cmd.TrackId, cmd.StreamId)
	if err != nil {
		return nil, errors.Wrap(err, "could not get the remote stream playlist")
	}

	return body, nil
}
