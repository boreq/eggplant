package domain

import (
	"encoding/hex"
	"errors"
)

type Thumbnail struct {
	id     ThumbnailId
	fileId FileId
}

func NewThumbnail(id ThumbnailId, fileId FileId) Thumbnail {
	return Thumbnail{id: id, fileId: fileId}
}

func (t Thumbnail) Id() ThumbnailId {
	return t.id
}

func (t Thumbnail) FileId() FileId {
	return t.fileId
}

type ThumbnailId struct {
	value string
}

func NewThumbnailIdFromString(s string) (ThumbnailId, error) {
	if s == "" {
		return ThumbnailId{}, errors.New("thumbnail id must not be empty")
	}
	if _, err := hex.DecodeString(s); err != nil {
		return ThumbnailId{}, errors.New("thumbnail id must be a hex string")
	}
	return ThumbnailId{value: s}, nil
}

func NewThumbnailId(parents []AlbumId, name FileName) (ThumbnailId, error) {
	return NewThumbnailIdFromString(hash(parentsAsString(parents) + name.value))
}

func (id ThumbnailId) String() string {
	return id.value
}
