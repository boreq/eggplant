package auth

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/boreq/errors"
)

func generateCryptoString(numBytes int) (string, error) {
	b := make([]byte, numBytes)
	if _, err := rand.Read(b); err != nil {
		return "", errors.Wrap(err, "could not read random bytes")
	}
	return hex.EncodeToString(b), nil
}
