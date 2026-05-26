package music

import (
	"github.com/boreq/eggplant/domain/music"
	"github.com/boreq/eggplant/domain/music/library"
	"github.com/boreq/errors"
)

type GetAlbum struct {
	Id music.AlbumId
}

type GetAlbumHandler struct {
	libraryRepository LibraryRepository
}

func NewGetAlbumHandler(repo LibraryRepository) *GetAlbumHandler {
	return &GetAlbumHandler{libraryRepository: repo}
}

func (h *GetAlbumHandler) Execute(accessCtx library.AccessContext, cmd GetAlbum) (music.Album, error) {
	lib, err := h.libraryRepository.Get()
	if err != nil {
		return music.Album{}, errors.Wrap(err, "could not get the library")
	}

	album, err := lib.GetAlbum(accessCtx, cmd.Id)
	if err != nil {
		return music.Album{}, errors.Wrap(err, "could not get the album")
	}

	return album, nil
}
