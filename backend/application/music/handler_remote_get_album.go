package music

import (
	"context"

	"github.com/boreq/eggplant/application/accessctx"
	"github.com/boreq/eggplant/domain/music"
	"github.com/boreq/eggplant/domain/remote"
	"github.com/boreq/errors"
)

type RemoteGetAlbum struct {
	instanceId remote.RemoteInstanceID
	albumId    music.AlbumId
}

func NewRemoteGetAlbum(instanceId remote.RemoteInstanceID, albumId music.AlbumId) RemoteGetAlbum {
	return RemoteGetAlbum{
		instanceId: instanceId,
		albumId:    albumId,
	}
}

func (r RemoteGetAlbum) InstanceId() remote.RemoteInstanceID {
	return r.instanceId
}

func (r RemoteGetAlbum) AlbumId() music.AlbumId {
	return r.albumId
}

type RemoteGetAlbumHandler struct {
	remoteLibrary RemoteLibrary
}

func NewRemoteGetAlbumHandler(remoteLibrary RemoteLibrary) *RemoteGetAlbumHandler {
	return &RemoteGetAlbumHandler{remoteLibrary: remoteLibrary}
}

func (h *RemoteGetAlbumHandler) Execute(ctx context.Context, accessCtx accessctx.AccessContext, cmd RemoteGetAlbum) (music.Album, error) {
	if !accessCtx.Can(accessctx.PermissionSeeRemoteLibraryContent) {
		return music.Album{}, accessctx.ErrPermissionDenied
	}

	album, err := h.remoteLibrary.GetAlbum(ctx, cmd.InstanceId(), cmd.AlbumId())
	if err != nil {
		return music.Album{}, errors.Wrap(err, "could not get the remote album")
	}

	return album, nil
}
