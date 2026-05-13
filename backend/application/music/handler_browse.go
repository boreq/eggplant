package music

import (
	"github.com/boreq/eggplant/domain"
	"github.com/boreq/errors"
)

type Browse struct {
	Ids        []domain.AlbumId
	PublicOnly bool
}

type BrowseHandler struct {
	library Library
}

func NewBrowseHandler(library Library) *BrowseHandler {
	return &BrowseHandler{
		library: library,
	}
}

func (h *BrowseHandler) Execute(cmd Browse) (domain.Album, error) {
	album, err := h.library.Browse(cmd.Ids, cmd.PublicOnly)
	if err != nil {
		return domain.Album{}, errors.Wrap(err, "could not browse the album")
	}

	if len(cmd.Ids) > 0 && len(album.Albums()) == 0 && len(album.Tracks()) == 0 {
		return domain.Album{}, ErrForbidden
	}

	return album, nil
}
