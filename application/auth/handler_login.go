package auth

import (
	"time"

	"github.com/boreq/errors"
)

type Login struct {
	Username string
	Password string
}

type LoginHandler struct {
	passwordHasher       PasswordHasher
	transactionProvider  TransactionProvider
	accessTokenGenerator AccessTokenGenerator
}

func NewLoginHandler(
	passwordHasher PasswordHasher,
	transactionProvider TransactionProvider,
	accessTokenGenerator AccessTokenGenerator,
) *LoginHandler {
	return &LoginHandler{
		passwordHasher:       passwordHasher,
		transactionProvider:  transactionProvider,
		accessTokenGenerator: accessTokenGenerator,
	}
}

func (h *LoginHandler) Execute(cmd Login) (AccessToken, error) {
	if err := validate(cmd.Username, cmd.Password); err != nil {
		return "", errors.Wrap(ErrUnauthorized, "invalid parameters")
	}

	var token AccessToken

	if err := h.transactionProvider.Write(func(r *TransactableRepositories) error {
		u, err := r.Users.Get(cmd.Username)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return errors.Wrap(ErrUnauthorized, "user not found")
			}
			return errors.Wrap(err, "could not get the user")
		}

		if err := h.passwordHasher.Compare(u.Password, cmd.Password); err != nil {
			return errors.Wrap(ErrUnauthorized, "invalid password")
		}

		t, err := h.accessTokenGenerator.Generate(cmd.Username)
		if err != nil {
			return errors.Wrap(err, "could not create an access token")
		}
		token = t

		s := Session{
			Token:    t,
			LastSeen: time.Now(),
		}

		u.Sessions = append(u.Sessions, s)

		return r.Users.Put(*u)
	}); err != nil {
		return "", errors.Wrap(err, "transaction failed")
	}

	return token, nil
}
