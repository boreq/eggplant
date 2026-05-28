package auth

import (
	authdomain "github.com/boreq/eggplant/domain/auth"
	"github.com/boreq/errors"
)

type Remove struct {
	Username authdomain.Username
}

type RemoveHandler struct {
	transactionProvider TransactionProvider
}

func NewRemoveHandler(transactionProvider TransactionProvider) *RemoveHandler {
	return &RemoveHandler{
		transactionProvider: transactionProvider,
	}
}

func (h *RemoveHandler) Execute(accessCtx AuthenticatedAccessContext, cmd Remove) error {
	if !accessCtx.Can(PermissionManageUsers) {
		return ErrPermissionDenied
	}

	if accessCtx.Username() == cmd.Username {
		return ErrCannotRemoveSelf
	}

	return h.transactionProvider.Write(func(r *TransactableRepositories) error {
		u, err := r.Users.Get(cmd.Username)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return nil
			}
			return errors.Wrap(err, "could not get the user")
		}

		for _, s := range u.Sessions() {
			if err := r.SessionTokens.Remove(s.Token()); err != nil {
				return errors.Wrap(err, "could not remove the session token")
			}
		}

		return r.Users.Remove(cmd.Username)
	})
}
