package auth

import (
	"time"

	"github.com/pkg/errors"
)

type RegisterInitial struct {
	Username string
	Password string
}

type RegisterInitialHandler struct {
	passwordHasher      PasswordHasher
	transactionProvider TransactionProvider
}

func NewRegisterInitialHandler(
	passwordHasher PasswordHasher,
	transactionProvider TransactionProvider,
) *RegisterInitialHandler {
	return &RegisterInitialHandler{
		passwordHasher:      passwordHasher,
		transactionProvider: transactionProvider,
	}
}

func (h *RegisterInitialHandler) Execute(cmd RegisterInitial) error {
	if err := validate(cmd.Username, cmd.Password); err != nil {
		return errors.Wrap(err, "invalid parameters")
	}

	passwordHash, err := h.passwordHasher.Hash(cmd.Password)
	if err != nil {
		return errors.Wrap(err, "hashing the password failed")
	}

	u := User{
		Username:      cmd.Username,
		Password:      passwordHash,
		Administrator: true,
		Created:       time.Now(),
		LastSeen:      time.Now(),
	}

	if err := h.transactionProvider.Write(func(r *TransactableRepositories) error {
		n, err := r.Users.Count()
		if err != nil {
			return errors.Wrap(err, "could not get a count")
		}
		if n != 0 {
			return errors.New("there are existing users")
		}
		return r.Users.Put(u)
	}); err != nil {
		return errors.Wrap(err, "transaction failed")
	}

	return nil
}
