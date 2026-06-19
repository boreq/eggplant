package remote

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"

	"github.com/boreq/errors"
)

const (
	pairingTokenBytes = 256 / 8
	authTokenBytes    = 256 / 8
)

type PairingToken struct {
	token
}

func NewPairingToken() (PairingToken, error) {
	t, err := newToken(pairingTokenBytes)
	if err != nil {
		return PairingToken{}, errors.Wrap(err, "could not generate the pairing token")
	}
	return PairingToken{t}, nil
}

func NewPairingTokenFromBytes(b []byte) (PairingToken, error) {
	t, err := newTokenFromBytes(b)
	if err != nil {
		return PairingToken{}, errors.Wrap(err, "invalid pairing token")
	}
	return PairingToken{t}, nil
}

func (t PairingToken) Hash() HashedPairingToken {
	return HashedPairingToken{t.hashed()}
}

func (t PairingToken) Equal(o PairingToken) bool {
	return bytes.Equal(t.value, o.value)
}

type HashedPairingToken struct {
	hashedToken
}

func NewHashedPairingTokenFromBytes(b []byte) (HashedPairingToken, error) {
	t, err := newHashedTokenFromBytes(b)
	if err != nil {
		return HashedPairingToken{}, errors.Wrap(err, "invalid hashed pairing token")
	}
	return HashedPairingToken{t}, nil
}

func (t HashedPairingToken) Equal(o HashedPairingToken) bool {
	return bytes.Equal(t.value, o.value)
}

type AuthToken struct {
	token
}

func NewAuthToken() (AuthToken, error) {
	t, err := newToken(authTokenBytes)
	if err != nil {
		return AuthToken{}, errors.Wrap(err, "could not generate the auth token")
	}
	return AuthToken{t}, nil
}

func NewAuthTokenFromBytes(b []byte) (AuthToken, error) {
	t, err := newTokenFromBytes(b)
	if err != nil {
		return AuthToken{}, errors.Wrap(err, "invalid auth token")
	}
	return AuthToken{t}, nil
}

func (t AuthToken) Hash() HashedAuthToken {
	return HashedAuthToken{t.hashed()}
}

func (t AuthToken) Equal(o AuthToken) bool {
	return bytes.Equal(t.value, o.value)
}

type HashedAuthToken struct {
	hashedToken
}

func NewHashedAuthTokenFromBytes(b []byte) (HashedAuthToken, error) {
	t, err := newHashedTokenFromBytes(b)
	if err != nil {
		return HashedAuthToken{}, errors.Wrap(err, "invalid hashed auth token")
	}
	return HashedAuthToken{t}, nil
}

type token struct {
	value []byte
}

func newToken(numBytes int) (token, error) {
	b := make([]byte, numBytes)
	if _, err := rand.Read(b); err != nil {
		return token{}, errors.Wrap(err, "could not read random bytes")
	}
	return token{value: b}, nil
}

func newTokenFromBytes(b []byte) (token, error) {
	if len(b) == 0 {
		return token{}, errors.New("token must not be empty")
	}
	return token{value: b}, nil
}

func (t token) Bytes() []byte {
	return t.value
}

func (t token) hashed() hashedToken {
	sum := sha256.Sum256(t.value)
	return hashedToken{value: sum[:]}
}

type hashedToken struct {
	value []byte
}

func newHashedTokenFromBytes(b []byte) (hashedToken, error) {
	if len(b) == 0 {
		return hashedToken{}, errors.New("hashed token must not be empty")
	}
	return hashedToken{value: b}, nil
}

func (t hashedToken) Bytes() []byte {
	return t.value
}
