package auth

import (
	"github.com/boreq/eggplant/application/accessctx"
	authdomain "github.com/boreq/eggplant/domain/auth"
	"github.com/boreq/errors"
)

type SetPassword struct {
	username authdomain.Username
	password authdomain.Password
}

func NewSetPassword(username authdomain.Username, password authdomain.Password) SetPassword {
	return SetPassword{
		username: username,
		password: password,
	}
}

func (s SetPassword) Username() authdomain.Username {
	return s.username
}

func (s SetPassword) Password() authdomain.Password {
	return s.password
}

type SetPasswordHandler struct {
	passwordHasher      PasswordHasher
	transactionProvider TransactionProvider
}

func NewSetPasswordHandler(
	passwordHasher PasswordHasher,
	transactionProvider TransactionProvider,
) *SetPasswordHandler {
	return &SetPasswordHandler{
		passwordHasher:      passwordHasher,
		transactionProvider: transactionProvider,
	}
}

func (h *SetPasswordHandler) Execute(accessCtx accessctx.AccessContext, cmd SetPassword) error {
	if !accessCtx.Can(accessctx.PermissionManageUsers) {
		return accessctx.ErrPermissionDenied
	}

	passwordHash, err := h.passwordHasher.Hash(cmd.Password())
	if err != nil {
		return errors.Wrap(err, "hashing the password failed")
	}

	return h.transactionProvider.Write(func(r *TransactableRepositories) error {
		u, err := r.Users.Get(cmd.Username())
		if err != nil {
			return errors.Wrap(err, "could not get the user")
		}

		u.SetPassword(passwordHash)

		return r.Users.Put(*u)
	})
}
