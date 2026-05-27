package auth

import (
	authdomain "github.com/boreq/eggplant/domain/auth"
	"github.com/boreq/errors"
)

type GetCurrentUserHandler struct {
	transactionProvider TransactionProvider
}

func NewGetCurrentUserHandler(transactionProvider TransactionProvider) *GetCurrentUserHandler {
	return &GetCurrentUserHandler{
		transactionProvider: transactionProvider,
	}
}

func (h *GetCurrentUserHandler) Execute(accessCtx AuthenticatedAccessContext) (authdomain.User, error) {
	var user authdomain.User
	if err := h.transactionProvider.Read(func(r *TransactableRepositories) error {
		u, err := r.Users.Get(accessCtx.Username())
		if err != nil {
			return errors.Wrap(err, "could not get the user")
		}
		user = *u
		return nil
	}); err != nil {
		return authdomain.User{}, errors.Wrap(err, "transaction failed")
	}
	return user, nil
}
