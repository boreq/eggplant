package auth

import (
	"github.com/boreq/eggplant/application/accessctx"
	authdomain "github.com/boreq/eggplant/domain/auth"
	"github.com/boreq/errors"
)

type ListHandler struct {
	transactionProvider TransactionProvider
}

func NewListHandler(transactionProvider TransactionProvider) *ListHandler {
	return &ListHandler{
		transactionProvider: transactionProvider,
	}
}

func (h *ListHandler) Execute(accessCtx accessctx.AccessContext) ([]authdomain.User, error) {
	if !accessCtx.Can(accessctx.PermissionManageUsers) {
		return nil, accessctx.ErrPermissionDenied
	}

	var users []authdomain.User
	if err := h.transactionProvider.Read(func(r *TransactableRepositories) error {
		u, err := r.Users.List()
		if err != nil {
			return errors.Wrap(err, "could not list the users")
		}
		users = u
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}
	return users, nil
}
