package auth

import "github.com/boreq/errors"

type SetPassword struct {
	Username string
	Password string
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

func (h *SetPasswordHandler) Execute(cmd SetPassword) error {
	if err := validate(cmd.Username, cmd.Password); err != nil {
		return errors.Wrap(err, "invalid parameters")
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

		u.Password = passwordHash

		return r.Users.Put(*u)
	})
}
