package music

import (
	"context"

	"github.com/boreq/eggplant/application/accessctx"
	"github.com/boreq/eggplant/domain/music"
	"github.com/boreq/eggplant/internal/logging"
	"github.com/boreq/errors"
)

type GetRootAlbumHandler struct {
	libraryRepository LibraryRepository
	remoteLibrary     RemoteLibrary
	log               logging.Logger
}

func NewGetRootAlbumHandler(repo LibraryRepository, remoteLibrary RemoteLibrary) *GetRootAlbumHandler {
	return &GetRootAlbumHandler{
		libraryRepository: repo,
		remoteLibrary:     remoteLibrary,
		log:               logging.New("music.GetRootAlbumHandler"),
	}
}

type GetRootAlbum struct {
}

func (h *GetRootAlbumHandler) Execute(ctx context.Context, accessCtx accessctx.AccessContext, cmd GetRootAlbum) (music.RootAlbum, error) {
	lib, err := h.libraryRepository.Get()
	if err != nil {
		return music.RootAlbum{}, errors.Wrap(err, "could not get the library")
	}

	album, err := lib.GetRootAlbum(accessCtx)
	if err != nil {
		return music.RootAlbum{}, errors.Wrap(err, "could not get the root album")
	}

	if !accessCtx.Can(accessctx.PermissionSeeRemoteLibraryContent) {
		return album, nil
	}

	remoteAlbums, err := h.remoteLibrary.GetRootAlbums(ctx)
	if err != nil {
		h.log.Error("could not get remote root albums", "err", err)
		return album, nil
	}

	joined, err := music.MergeRootAlbums(album, remoteAlbums)
	if err != nil {
		return music.RootAlbum{}, errors.Wrap(err, "could not build the joined root album")
	}

	return joined, nil
}
