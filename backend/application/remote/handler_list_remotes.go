package remote

import (
	"github.com/boreq/eggplant/application/accessctx"
	remotedomain "github.com/boreq/eggplant/domain/remote"
	"github.com/boreq/errors"
)

type ListRemotesHandler struct {
	transactionProvider TransactionProvider
}

func NewListRemotesHandler(transactionProvider TransactionProvider) *ListRemotesHandler {
	return &ListRemotesHandler{transactionProvider: transactionProvider}
}

func (h *ListRemotesHandler) Execute(accessCtx accessctx.AccessContext) ([]*remotedomain.RemoteInstance, error) {
	if !accessCtx.Can(accessctx.PermissionManageRemotes) {
		return nil, accessctx.ErrPermissionDenied
	}

	var instances []*remotedomain.RemoteInstance

	if err := h.transactionProvider.Read(func(r *TransactableRepositories) error {
		all, err := r.RemoteInstances.GetAll()
		if err != nil {
			return errors.Wrap(err, "could not get the remote instances")
		}
		instances = all
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}

	return instances, nil
}
