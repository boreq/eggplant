package music

import (
	"github.com/boreq/eggplant/domain"
	"github.com/boreq/errors"
)

type GetAlbum struct {
	Id         domain.AlbumId
	PublicOnly bool
}

type GetAlbumHandler struct {
	libraryRepository LibraryRepository
}

func NewGetAlbumHandler(repo LibraryRepository) *GetAlbumHandler {
	return &GetAlbumHandler{libraryRepository: repo}
}

func (h *GetAlbumHandler) Execute(cmd GetAlbum) (domain.Album, error) {
	lib, err := h.libraryRepository.Get()
	if err != nil {
		return domain.Album{}, errors.Wrap(err, "could not get the library")
	}
	a, err := lib.GetAlbum(cmd.Id, cmd.PublicOnly)
	if err != nil {
		return domain.Album{}, errors.Wrap(err, "could not get the album")
	}
	return a, nil
}
