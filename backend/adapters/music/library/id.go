package library

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	musiclib "github.com/boreq/eggplant/application/music/library"
	"github.com/boreq/eggplant/domain"
	"github.com/boreq/errors"
)

type idGenerator struct {
}

func NewIdGenerator() musiclib.IdGenerator {
	return idGenerator{}
}

func (idGenerator) AlbumId(parents []domain.AlbumId, title string) (domain.AlbumId, error) {
	h, err := shortHash(parentsAsString(parents) + title)
	if err != nil {
		return domain.AlbumId{}, errors.Wrap(err, "hashing failed")
	}
	return domain.NewAlbumId(h)
}

func (idGenerator) TrackId(parents []domain.AlbumId, title string) (domain.TrackId, error) {
	h, err := shortHash(parentsAsString(parents) + title)
	if err != nil {
		return domain.TrackId{}, errors.Wrap(err, "hashing failed")
	}
	return domain.NewTrackId(h)
}

func (idGenerator) FileId(path string) (domain.FileId, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return domain.FileId{}, errors.Wrap(err, "os stat failed")
	}
	s := fmt.Sprintf("%s-%d-%d", path, fileInfo.Size(), fileInfo.ModTime().Unix())
	h, err := longHash(s)
	if err != nil {
		return domain.FileId{}, errors.Wrap(err, "hashing failed")
	}
	return domain.NewFileId(h)
}

func parentsAsString(parents []domain.AlbumId) string {
	var s string
	for _, parent := range parents {
		s += parent.String()
	}
	return s
}

func shortHash(s string) (string, error) {
	sum, err := longHash(s)
	if err != nil {
		return "", errors.Wrap(err, "hashing failed")
	}
	return sum[:20], nil
}

func longHash(s string) (string, error) {
	buf := bytes.NewBuffer([]byte(s))
	hasher := sha256.New()
	if _, err := io.Copy(hasher, buf); err != nil {
		return "", err
	}
	var sum []byte
	sum = hasher.Sum(sum)
	return hex.EncodeToString(sum), nil
}
