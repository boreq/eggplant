package music

import (
	"context"

	"github.com/boreq/eggplant/domain"
	"github.com/boreq/errors"
)

type ThumbnailHandler struct {
	thumbnailStore ThumbnailStore
}

func NewThumbnailHandler(thumbnailStore ThumbnailStore) *ThumbnailHandler {
	return &ThumbnailHandler{
		thumbnailStore: thumbnailStore,
	}
}

func (h *ThumbnailHandler) Execute(ctx context.Context, id domain.ThumbnailId) (domain.ConvertedFile, error) {
	p, err := h.thumbnailStore.GetConvertedFile(ctx, id)
	if err != nil {
		return domain.ConvertedFile{}, errors.Wrap(err, "could not get the thumbnail")
	}
	return p, nil
}
