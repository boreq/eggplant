package music

import (
	"context"
	"io"

	"github.com/boreq/eggplant/application/accessctx"
	"github.com/boreq/eggplant/domain/music"
	"github.com/boreq/eggplant/domain/remote"
	"github.com/boreq/errors"
)

type RemoteGetThumbnail struct {
	instanceId  remote.RemoteInstanceID
	thumbnailId music.ThumbnailId
}

func NewRemoteGetThumbnail(instanceId remote.RemoteInstanceID, thumbnailId music.ThumbnailId) RemoteGetThumbnail {
	return RemoteGetThumbnail{
		instanceId:  instanceId,
		thumbnailId: thumbnailId,
	}
}

func (r RemoteGetThumbnail) InstanceId() remote.RemoteInstanceID {
	return r.instanceId
}

func (r RemoteGetThumbnail) ThumbnailId() music.ThumbnailId {
	return r.thumbnailId
}

type RemoteGetThumbnailHandler struct {
	client RemoteLibrary
}

func NewRemoteGetThumbnailHandler(client RemoteLibrary) *RemoteGetThumbnailHandler {
	return &RemoteGetThumbnailHandler{client: client}
}

func (h *RemoteGetThumbnailHandler) Execute(ctx context.Context, accessCtx accessctx.AccessContext, cmd RemoteGetThumbnail) (io.ReadCloser, error) {
	if !accessCtx.Can(accessctx.PermissionSeeRemoteLibraryContent) {
		return nil, accessctx.ErrPermissionDenied
	}

	body, err := h.client.GetThumbnail(ctx, cmd.InstanceId(), cmd.ThumbnailId())
	if err != nil {
		return nil, errors.Wrap(err, "could not get the remote thumbnail")
	}

	return body, nil
}
