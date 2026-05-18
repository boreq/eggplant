package domain

import (
	"crypto/sha256"

	"github.com/boreq/errors"
)

// FileId must never be exposed outside the backend: it is a hash of the
// full file path, which carries more information than the album and track
// titles a caller already sees (the on-disk directory layout, exact
// filenames, extensions, the music root location).
type FileId struct {
	value string
}

func NewFileId(path FilePath) (FileId, error) {
	if path.value == "" {
		return FileId{}, errors.New("file path must not be empty")
	}
	sum := sha256.Sum256([]byte(path.value))
	return FileId{value: encodeCrockford(sum[:])}, nil
}

func (f FileId) String() string {
	return f.value
}
