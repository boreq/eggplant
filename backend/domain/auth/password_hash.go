package auth

import "github.com/boreq/errors"

type PasswordHash struct {
	value []byte
}

func NewPasswordHash(b []byte) (PasswordHash, error) {
	if len(b) == 0 {
		return PasswordHash{}, errors.New("password hash must not be empty")
	}
	return PasswordHash{value: b}, nil
}

func (h PasswordHash) Bytes() []byte {
	return h.value
}
