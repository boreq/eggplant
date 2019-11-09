package auth

import (
	"github.com/boreq/eggplant/application/auth"
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

func (p *BcryptPasswordHasher) Hash(password string) (auth.PasswordHash, error) {
	return bcrypt.GenerateFromPassword([]byte(password), p.cost)
}

func (p *BcryptPasswordHasher) Compare(hashedPassword auth.PasswordHash, password string) error {
	return bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
}
