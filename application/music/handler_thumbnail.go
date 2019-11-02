package music

import "github.com/boreq/errors"

type ThumbnailHandler struct {
	thumbnailStore ThumbnailStore
}

func NewThumbnailHandler(thumbnailStore ThumbnailStore) *ThumbnailHandler {
	return &ThumbnailHandler{
		thumbnailStore: thumbnailStore,
	}
}

func (h *ThumbnailHandler) Execute(id string) (string, error) {
	p, err := h.thumbnailStore.GetFilePath(id)
	if err != nil {
		return "", errors.Wrap(err, "could not get the thumbnail path")
	}
	return p, nil
}
