package auth

import (
	"fmt"

	"github.com/boreq/errors"
)

const maxPasswordLen = 10000
const randomPasswordBytes = 256 / 8

type Password struct {
	value string
}

func NewPasswordFromString(s string) (Password, error) {
	if s == "" {
		return Password{}, errors.New("password can't be empty")
	}
	if len(s) > maxPasswordLen {
		return Password{}, fmt.Errorf("password length can't exceed %d characters", maxPasswordLen)
	}
	return Password{value: s}, nil
}

func GenerateRandomPassword() (Password, error) {
	s, err := generateCryptoString(randomPasswordBytes)
	if err != nil {
		return Password{}, errors.Wrap(err, "could not generate the password")
	}
	return Password{value: s}, nil
}

func (p Password) String() string {
	return p.value
}
