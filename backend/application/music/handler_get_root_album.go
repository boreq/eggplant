package music

import (
	"github.com/boreq/eggplant/domain"
	"github.com/boreq/eggplant/domain/library"
	"github.com/boreq/errors"
)

type GetRootAlbumHandler struct {
	libraryRepository LibraryRepository
}

func NewGetRootAlbumHandler(repo LibraryRepository) *GetRootAlbumHandler {
	return &GetRootAlbumHandler{libraryRepository: repo}
}

func (h *GetRootAlbumHandler) Execute(accessCtx library.AccessContext) (domain.RootAlbum, error) {
	lib, err := h.libraryRepository.Get()
	if err != nil {
		return domain.RootAlbum{}, errors.Wrap(err, "could not get the library")
	}

	album, err := lib.GetRootAlbum(accessCtx)
	if err != nil {
		return domain.RootAlbum{}, errors.Wrap(err, "could not get the root album")
	}

	return album, nil
}
