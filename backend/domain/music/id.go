package music

import (
	"crypto/sha256"
	"encoding/base32"
	"fmt"

	"github.com/boreq/errors"
)

const (
	crockfordBase32Alphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"
	idHashBytes             = 10
)

var crockfordBase32 = base32.NewEncoding(crockfordBase32Alphabet).WithPadding(base32.NoPadding)

type idForHumans struct {
	value string
}

func newIdForHumans(parents []AlbumId, last fmt.Stringer) idForHumans {
	s := parentsAsString(parents) + last.String()
	sum := sha256.Sum256([]byte(s))
	return idForHumans{value: encodeCrockford(sum[:idHashBytes])}
}

func newIdForHumansFromString(s string) (idForHumans, error) {
	if s == "" {
		return idForHumans{}, errors.New("id must not be empty")
	}
	decoded, err := decodeCrockford(s)
	if err != nil {
		return idForHumans{}, errors.Wrap(err, "id must be a Crockford base32 string")
	}
	if len(decoded) != idHashBytes {
		return idForHumans{}, errors.New("id has wrong length")
	}
	return idForHumans{value: s}, nil
}

func (id idForHumans) String() string {
	return id.value
}

func parentsAsString(parents []AlbumId) string {
	var s string
	for _, p := range parents {
		s += p.String()
	}
	return s
}

func encodeCrockford(b []byte) string {
	return crockfordBase32.EncodeToString(b)
}

func decodeCrockford(s string) ([]byte, error) {
	return crockfordBase32.DecodeString(s)
}
