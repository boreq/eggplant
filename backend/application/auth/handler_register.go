package auth

import (
	"time"

	"github.com/boreq/eggplant/application/accessctx"
	authdomain "github.com/boreq/eggplant/domain/auth"
	"github.com/boreq/errors"
)

type Register struct {
	username authdomain.Username
	password authdomain.Password
	token    authdomain.InvitationToken
}

func NewRegister(username authdomain.Username, password authdomain.Password, token authdomain.InvitationToken) Register {
	return Register{
		username: username,
		password: password,
		token:    token,
	}
}

func (r Register) Username() authdomain.Username {
	return r.username
}

func (r Register) Password() authdomain.Password {
	return r.password
}

func (r Register) Token() authdomain.InvitationToken {
	return r.token
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

func (h *RegisterHandler) Execute(accessCtx accessctx.AccessContext, cmd Register) error {
	if _, ok := accessCtx.(accessctx.UserAccessContext); ok {
		return ErrAlreadyAuthenticated
	}

	passwordHash, err := h.passwordHasher.Hash(cmd.Password())
	if err != nil {
		return errors.Wrap(err, "hashing the password failed")
	}

	u := authdomain.NewUser(cmd.Username(), passwordHash, false, time.Now())

	if err := h.transactionProvider.Write(func(r *TransactableRepositories) error {
		if _, err := r.Invitations.Get(cmd.Token()); err != nil {
			return errors.Wrap(err, "could not get the invitation")
		}

		if err := r.Invitations.Remove(cmd.Token()); err != nil {
			return errors.Wrap(err, "could not remove the invitation")
		}

		if _, err := r.Users.Get(cmd.Username()); err != nil {
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
