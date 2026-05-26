package auth

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/boreq/errors"
)

const accessTokenSecretBytes = 256 / 8
const accessTokenSeparator = "-"

type AccessToken struct {
	value string
}

func NewAccessToken(username Username) (AccessToken, error) {
	secret, err := generateCryptoString(accessTokenSecretBytes)
	if err != nil {
		return AccessToken{}, errors.Wrap(err, "could not generate the secret")
	}
	encodedUsername := hex.EncodeToString([]byte(username.String()))
	return AccessToken{
		value: fmt.Sprintf("%s%s%s", secret, accessTokenSeparator, encodedUsername),
	}, nil
}

func NewAccessTokenFromString(s string) (AccessToken, error) {
	if s == "" {
		return AccessToken{}, errors.New("access token must not be empty")
	}
	return AccessToken{value: s}, nil
}

func (t AccessToken) Username() (Username, error) {
	parts := strings.Split(t.value, accessTokenSeparator)
	if len(parts) != 2 {
		return Username{}, errors.New("malformed token")
	}
	h, err := hex.DecodeString(parts[1])
	if err != nil {
		return Username{}, errors.Wrap(err, "hex decoding failed")
	}
	return NewUsernameFromString(string(h))
}

func (t AccessToken) String() string {
	return t.value
}
