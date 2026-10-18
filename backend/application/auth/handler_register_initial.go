package auth

import (
	"time"

	authdomain "github.com/boreq/eggplant/domain/auth"
	"github.com/pkg/errors"
)

type RegisterInitial struct {
	username authdomain.Username
	password authdomain.Password
}

func NewRegisterInitial(username authdomain.Username, password authdomain.Password) RegisterInitial {
	return RegisterInitial{
		username: username,
		password: password,
	}
}

func (r RegisterInitial) Username() authdomain.Username {
	return r.username
}

func (r RegisterInitial) Password() authdomain.Password {
	return r.password
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
	passwordHash, err := h.passwordHasher.Hash(cmd.Password())
	if err != nil {
		return errors.Wrap(err, "hashing the password failed")
	}

	u := authdomain.NewUser(cmd.Username(), passwordHash, true, time.Now())

	if err := h.transactionProvider.Write(func(r *TransactableRepositories) error {
		n, err := r.Users.Count()
		if err != nil {
			return errors.Wrap(err, "could not get a count")
		}
		if n != 0 {
			return ErrUsersAlreadyExist
		}
		return r.Users.Put(u)
	}); err != nil {
		return errors.Wrap(err, "transaction failed")
	}

	return nil
}
