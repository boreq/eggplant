package auth

import (
	"time"

	"github.com/boreq/eggplant/application/accessctx"
	authdomain "github.com/boreq/eggplant/domain/auth"
	"github.com/boreq/errors"
)

type Login struct {
	Username authdomain.Username
	Password authdomain.Password
}

type LoginHandler struct {
	passwordHasher      PasswordHasher
	transactionProvider TransactionProvider
}

func NewLoginHandler(
	passwordHasher PasswordHasher,
	transactionProvider TransactionProvider,
) *LoginHandler {
	return &LoginHandler{
		passwordHasher:      passwordHasher,
		transactionProvider: transactionProvider,
	}
}

func (h *LoginHandler) Execute(accessCtx accessctx.AccessContext, cmd Login) (authdomain.AccessToken, error) {
	if _, ok := accessCtx.(accessctx.UserAccessContext); ok {
		return authdomain.AccessToken{}, ErrAlreadyAuthenticated
	}

	var token authdomain.AccessToken

	if err := h.transactionProvider.Write(func(r *TransactableRepositories) error {
		u, err := r.Users.Get(cmd.Username)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return errors.Wrap(ErrUnauthorized, "user not found")
			}
			return errors.Wrap(err, "could not get the user")
		}

		if err := h.passwordHasher.Compare(u.Password(), cmd.Password); err != nil {
			return errors.Wrap(ErrUnauthorized, "invalid password")
		}

		t, err := authdomain.NewAccessToken()
		if err != nil {
			return errors.Wrap(err, "could not create an access token")
		}
		token = t

		u.AddSession(authdomain.NewSession(t, time.Now()))

		if err := r.Users.Put(*u); err != nil {
			return errors.Wrap(err, "could not put the user")
		}

		return nil
	}); err != nil {
		return authdomain.AccessToken{}, errors.Wrap(err, "transaction failed")
	}

	return token, nil
}
