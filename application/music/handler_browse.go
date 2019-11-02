package music

import "github.com/boreq/errors"

type Browse struct {
	Ids        []AlbumId
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

func (h *BrowseHandler) Execute(cmd Browse) (Album, error) {
	album, err := h.library.Browse(cmd.Ids, cmd.PublicOnly)
	if err != nil {
		return Album{}, errors.Wrap(err, "could not browse the album")
	}
	return album, nil
}
