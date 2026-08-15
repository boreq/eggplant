package music

import (
	"github.com/boreq/eggplant/domain/music"
	"github.com/boreq/eggplant/domain/music/library"
	"github.com/boreq/errors"
)

type GetRootAlbumHandler struct {
	libraryRepository LibraryRepository
}

func NewGetRootAlbumHandler(repo LibraryRepository) *GetRootAlbumHandler {
	return &GetRootAlbumHandler{libraryRepository: repo}
}

type GetRootAlbum struct {
}

func (h *GetRootAlbumHandler) Execute(accessCtx library.AccessContext, cmd GetRootAlbum) (music.RootAlbum, error) {
	lib, err := h.libraryRepository.Get()
	if err != nil {
		return music.RootAlbum{}, errors.Wrap(err, "could not get the library")
	}

	album, err := lib.GetRootAlbum(accessCtx)
	if err != nil {
		return music.RootAlbum{}, errors.Wrap(err, "could not get the root album")
	}

	return album, nil
}
