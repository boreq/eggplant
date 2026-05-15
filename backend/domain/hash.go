package domain

import (
	"crypto/sha256"
	"encoding/hex"
)

const shortHashLength = 20

func shortHash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])[:shortHashLength]
}

func parentsAsString(parents []AlbumId) string {
	var s string
	for _, p := range parents {
		s += p.value
	}
	return s
}
