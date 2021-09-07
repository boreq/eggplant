package library

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"github.com/boreq/eggplant/application/music"
	"github.com/boreq/errors"
)

type idGenerator struct {
}

func NewIdGenerator() IdGenerator {
	return idGenerator{}
}

func (idGenerator) AlbumId(parents []music.AlbumId, title string) (music.AlbumId, error) {
	h, err := shortHash(parentsAsString(parents) + title)
	if err != nil {
		return "", errors.Wrap(err, "hashing failed")
	}
	return music.AlbumId(h), nil
}

func (idGenerator) TrackId(parents []music.AlbumId, title string) (music.TrackId, error) {
	h, err := shortHash(parentsAsString(parents) + title)
	if err != nil {
		return "", errors.Wrap(err, "hashing failed")
	}
	return music.TrackId(h), nil
}

func (idGenerator) FileId(path string) (music.FileId, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return "", errors.Wrap(err, "os stat failed")
	}
	s := fmt.Sprintf("%s-%d-%d", path, fileInfo.Size(), fileInfo.ModTime().Unix())
	h, err := longHash(s)
	if err != nil {
		return "", errors.Wrap(err, "hashing failed")
	}
	return music.FileId(h), nil
}

func parentsAsString(parents []music.AlbumId) (string) {
	var s string
	for _, parent := range parents {
		s += string(parent)
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
