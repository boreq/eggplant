package music

import "github.com/boreq/errors"

type BrowseHandler struct {
	library Library
}

func NewBrowseHandler(library Library) *BrowseHandler {
	return &BrowseHandler{
		library: library,
	}
}

func (h *BrowseHandler) Execute(ids []AlbumId) (Album, error) {
	album, err := h.library.Browse(ids)
	if err != nil {
		return Album{}, errors.Wrap(err, "could not browse the album")
	}
	return album, nil
}
