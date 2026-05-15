package music

import (
	"github.com/boreq/eggplant/domain"
	"github.com/boreq/eggplant/domain/library"
	"github.com/boreq/errors"
)

type GetAlbum struct {
	Id domain.AlbumId
}

type GetAlbumHandler struct {
	libraryRepository LibraryRepository
}

func NewGetAlbumHandler(repo LibraryRepository) *GetAlbumHandler {
	return &GetAlbumHandler{libraryRepository: repo}
}

func (h *GetAlbumHandler) Execute(accessCtx library.AccessContext, cmd GetAlbum) (domain.Album, error) {
	lib, err := h.libraryRepository.Get()
	if err != nil {
		return domain.Album{}, errors.Wrap(err, "could not get the library")
	}

	album, err := lib.GetAlbum(cmd.Id, accessCtx)
	if err != nil {
		return domain.Album{}, errors.Wrap(err, "could not get the album")
	}

	return album, nil
}
