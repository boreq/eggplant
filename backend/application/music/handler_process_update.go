package music

import (
	"github.com/boreq/eggplant/adapters/music/scanner"
	"github.com/boreq/errors"
)

type ProcessUpdateHandler struct {
	library Library
}

func NewProcessUpdateHandler(library Library) *ProcessUpdateHandler {
	return &ProcessUpdateHandler{
		library: library,
	}
}

func (h *ProcessUpdateHandler) Execute(scan scanner.Album) error {
	if err := h.library.Apply(scan); err != nil {
		return errors.Wrap(err, "could not apply the scan")
	}
	return nil
}
