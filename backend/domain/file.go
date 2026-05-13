package domain

import (
	"encoding/hex"
	"errors"
)

type FileId struct {
	value string
}

func NewFileId(s string) (FileId, error) {
	if s == "" {
		return FileId{}, errors.New("file id must not be empty")
	}
	if _, err := hex.DecodeString(s); err != nil {
		return FileId{}, errors.New("file id must be a hex string")
	}
	return FileId{value: s}, nil
}

func (id FileId) String() string {
	return id.value
}
