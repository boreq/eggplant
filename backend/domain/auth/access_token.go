package auth

import (
	"github.com/boreq/errors"
)

const accessTokenBytes = 256 / 8

type AccessToken struct {
	value string
}

func NewAccessToken() (AccessToken, error) {
	secret, err := generateCryptoString(accessTokenBytes)
	if err != nil {
		return AccessToken{}, errors.Wrap(err, "could not generate the secret")
	}
	return AccessToken{value: secret}, nil
}

func NewAccessTokenFromString(s string) (AccessToken, error) {
	if s == "" {
		return AccessToken{}, errors.New("access token must not be empty")
	}
	return AccessToken{value: s}, nil
}

func (t AccessToken) String() string {
	return t.value
}
