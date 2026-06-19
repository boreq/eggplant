package music

import (
	"crypto/sha256"
	"fmt"

	"github.com/boreq/eggplant/domain/crockford"
	"github.com/boreq/errors"
)

const idHashBytes = 10

type idForHumans struct {
	value string
}

func newIdForHumans(parents []AlbumId, last fmt.Stringer) idForHumans {
	s := parentsAsString(parents) + last.String()
	sum := sha256.Sum256([]byte(s))
	return idForHumans{value: crockford.Encode(sum[:idHashBytes])}
}

func newIdForHumansFromString(s string) (idForHumans, error) {
	if s == "" {
		return idForHumans{}, errors.New("id must not be empty")
	}
	decoded, err := crockford.Decode(s)
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
