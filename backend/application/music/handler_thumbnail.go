package music

import (
	"context"

	"github.com/boreq/eggplant/domain"
	"github.com/boreq/eggplant/domain/library"
	"github.com/boreq/errors"
)

type ThumbnailHandler struct {
	libraryRepository LibraryRepository
	thumbnailStore    ThumbnailStore
}

func NewThumbnailHandler(libraryRepository LibraryRepository, thumbnailStore ThumbnailStore) *ThumbnailHandler {
	return &ThumbnailHandler{
		libraryRepository: libraryRepository,
		thumbnailStore:    thumbnailStore,
	}
}

func (h *ThumbnailHandler) Execute(ctx context.Context, accessCtx library.AccessContext, id domain.ThumbnailId) (domain.ConvertedFile, error) {
	lib, err := h.libraryRepository.Get()
	if err != nil {
		return domain.ConvertedFile{}, errors.Wrap(err, "could not get the library")
	}

	thumbnail, err := lib.GetThumbnail(accessCtx, id)
	if err != nil {
		return domain.ConvertedFile{}, errors.Wrap(err, "could not get the thumbnail")
	}

	cf, err := h.thumbnailStore.GetConvertedFile(ctx, thumbnail.FileId())
	if err != nil {
		return domain.ConvertedFile{}, errors.Wrap(err, "could not get the converted file")
	}

	return cf, nil
}
