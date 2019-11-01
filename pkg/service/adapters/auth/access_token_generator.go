package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/boreq/eggplant/errors"
	"github.com/boreq/eggplant/pkg/service/application/auth"
)

const tokenLengthBytes = 256 / 8
const tokenSeparator = "-"

type CryptoAccessTokenGenerator struct {
}

func NewCryptoAccessTokenGenerator() *CryptoAccessTokenGenerator {
	return &CryptoAccessTokenGenerator{}
}

func (g *CryptoAccessTokenGenerator) Generate(username string) (auth.AccessToken, error) {
	cryptoString, err := generateCryptoString(tokenLengthBytes)
	if err != nil {
		return "", errors.Wrap(err, "could not create a crypto string")
	}

	encodedUsername := hex.EncodeToString([]byte(username))
	token := fmt.Sprintf("%s%s%s", cryptoString, tokenSeparator, encodedUsername)
	return auth.AccessToken(token), nil
}

func (g *CryptoAccessTokenGenerator) GetUsername(token auth.AccessToken) (string, error) {
	parts := strings.Split(string(token), tokenSeparator)
	if len(parts) != 2 {
		return "", errors.New("malformed token")
	}

	h, err := hex.DecodeString(parts[1])
	if err != nil {
		return "", errors.Wrap(err, "hex decoding failed")
	}

	return string(h), nil
}

// generateCryptoString creates a random string generated using a
// cryptographically secure random source. numBytes specifies the number of
// bytes retrieved from the random source, the length of the generated string
// will be different.
func generateCryptoString(numBytes int) (string, error) {
	token := make([]byte, numBytes)
	if _, err := rand.Read(token); err != nil {
		return "", errors.Wrap(err, "could not create a crypto string")
	}
	return hex.EncodeToString(token), nil
}
