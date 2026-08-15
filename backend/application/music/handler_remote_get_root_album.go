package music

import (
	"context"

	"github.com/boreq/eggplant/application/accessctx"
	"github.com/boreq/eggplant/domain/music"
	"github.com/boreq/eggplant/domain/remote"
	"github.com/boreq/errors"
)

type RemoteGetRootAlbum struct {
	InstanceId remote.RemoteInstanceID
}

type RemoteGetRootAlbumHandler struct {
	remoteLibrary RemoteLibrary
}

func NewRemoteGetRootAlbumHandler(remoteLibrary RemoteLibrary) *RemoteGetRootAlbumHandler {
	return &RemoteGetRootAlbumHandler{remoteLibrary: remoteLibrary}
}

func (h *RemoteGetRootAlbumHandler) Execute(ctx context.Context, accessCtx accessctx.AccessContext, cmd RemoteGetRootAlbum) (music.RootAlbum, error) {
	if !accessCtx.Can(accessctx.PermissionSeeRemoteLibraryContent) {
		return music.RootAlbum{}, accessctx.ErrPermissionDenied
	}

	album, err := h.remoteLibrary.GetRootAlbum(ctx, cmd.InstanceId)
	if err != nil {
		return music.RootAlbum{}, errors.Wrap(err, "could not get the remote root album")
	}

	return album, nil
}
