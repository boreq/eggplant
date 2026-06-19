package auth

import (
	"github.com/boreq/eggplant/application/accessctx"
	authdomain "github.com/boreq/eggplant/domain/auth"
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

func (h *RemoveHandler) Execute(accessCtx accessctx.UserAccessContext, cmd Remove) error {
	if !accessCtx.Can(accessctx.PermissionManageUsers) {
		return accessctx.ErrPermissionDenied
	}

	if accessCtx.Username() == cmd.Username {
		return ErrCannotRemoveSelf
	}

	return h.transactionProvider.Write(func(r *TransactableRepositories) error {
		return r.Users.Remove(cmd.Username)
	})
}
