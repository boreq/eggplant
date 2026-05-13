package domain

type Thumbnail struct {
	fileId FileId
}

func NewThumbnail(fileId FileId) Thumbnail {
	return Thumbnail{fileId: fileId}
}

func (t Thumbnail) FileId() FileId {
	return t.fileId
}
