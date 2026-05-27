package auth

import (
	authdomain "github.com/boreq/eggplant/domain/auth"
	"github.com/boreq/errors"
)

type SetPassword struct {
	Username authdomain.Username
	Password authdomain.Password
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

func (h *SetPasswordHandler) Execute(accessCtx AccessContext, cmd SetPassword) error {
	if !accessCtx.Can(PermissionManageUsers) {
		return ErrUnauthorized
	}

	passwordHash, err := h.passwordHasher.Hash(cmd.Password)
	if err != nil {
		return errors.Wrap(err, "hashing the password failed")
	}

	return h.transactionProvider.Write(func(r *TransactableRepositories) error {
		u, err := r.Users.Get(cmd.Username)
		if err != nil {
			return errors.Wrap(err, "could not get the user")
		}

		u.SetPassword(passwordHash)

		return r.Users.Put(*u)
	})
}
