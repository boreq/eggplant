package domain

import (
	"encoding/hex"
	"errors"
)

type Thumbnail struct {
	id ThumbnailId
}

func NewThumbnail(id ThumbnailId) Thumbnail {
	return Thumbnail{id: id}
}

func (t Thumbnail) Id() ThumbnailId {
	return t.id
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
	return NewThumbnailIdFromString(shortHash(parentsAsString(parents) + name.value))
}

func (id ThumbnailId) String() string {
	return id.value
}
