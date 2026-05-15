package domain

import "errors"

type FileId struct {
	value string
}

func NewFileId(path FilePath) (FileId, error) {
	if path.value == "" {
		return FileId{}, errors.New("file path must not be empty")
	}
	return FileId{value: hash(path.value)}, nil
}

func (f FileId) String() string {
	return f.value
}
