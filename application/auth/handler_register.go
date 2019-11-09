package auth

import (
	"time"

	"github.com/boreq/errors"
)

type Register struct {
	Username string
	Password string
	Token    InvitationToken
}

type RegisterHandler struct {
	passwordHasher      PasswordHasher
	transactionProvider TransactionProvider
}

func NewRegisterHandler(
	passwordHasher PasswordHasher,
	transactionProvider TransactionProvider,
) *RegisterHandler {
	return &RegisterHandler{
		passwordHasher:      passwordHasher,
		transactionProvider: transactionProvider,
	}
}

func (h *RegisterHandler) Execute(cmd Register) error {
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
		Administrator: false,
		Created:       time.Now(),
		LastSeen:      time.Now(),
	}

	if err := h.transactionProvider.Write(func(r *TransactableRepositories) error {
		if _, err := r.Invitations.Get(cmd.Token); err != nil {
			return errors.Wrap(err, "could not get the invitation")
		}

		if err := r.Invitations.Remove(cmd.Token); err != nil {
			return errors.Wrap(err, "could not remove the invitation")
		}

		if _, err := r.Users.Get(cmd.Username); err != nil {
			if !errors.Is(err, ErrNotFound) {
				return errors.Wrap(err, "could not get the user")
			}
		} else {
			return ErrUsernameTaken
		}

		return r.Users.Put(u)
	}); err != nil {
		return errors.Wrap(err, "transaction failed")
	}

	return nil
}
