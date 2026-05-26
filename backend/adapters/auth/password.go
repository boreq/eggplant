package auth

import (
	authdomain "github.com/boreq/eggplant/domain/auth"
	"github.com/boreq/errors"
	"golang.org/x/crypto/bcrypt"
)

type BcryptPasswordHasher struct {
	cost int
}

func NewBcryptPasswordHasher() *BcryptPasswordHasher {
	return &BcryptPasswordHasher{
		cost: 12,
	}
}

func (p *BcryptPasswordHasher) Hash(password authdomain.Password) (authdomain.PasswordHash, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password.String()), p.cost)
	if err != nil {
		return authdomain.PasswordHash{}, errors.Wrap(err, "could not hash the password")
	}
	return authdomain.NewPasswordHash(b)
}

func (p *BcryptPasswordHasher) Compare(hashedPassword authdomain.PasswordHash, password authdomain.Password) error {
	return bcrypt.CompareHashAndPassword(hashedPassword.Bytes(), []byte(password.String()))
}
