package music

import "github.com/boreq/errors"

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
	id idForHumans
}

func NewThumbnailId(parents []AlbumId, name FileName) (ThumbnailId, error) {
	return ThumbnailId{id: newIdForHumans(parents, name)}, nil
}

func NewThumbnailIdFromString(s string) (ThumbnailId, error) {
	id, err := newIdForHumansFromString(s)
	if err != nil {
		return ThumbnailId{}, errors.Wrap(err, "invalid thumbnail id")
	}
	return ThumbnailId{id: id}, nil
}

func (t ThumbnailId) String() string {
	return t.id.String()
}
