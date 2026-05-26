package auth

import (
	"time"

	authdomain "github.com/boreq/eggplant/domain/auth"
	"github.com/pkg/errors"
)

type RegisterInitial struct {
	Username authdomain.Username
	Password authdomain.Password
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
	passwordHash, err := h.passwordHasher.Hash(cmd.Password)
	if err != nil {
		return errors.Wrap(err, "hashing the password failed")
	}

	now := time.Now()
	u := authdomain.NewUser(cmd.Username, passwordHash, true, now, now, nil)

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
