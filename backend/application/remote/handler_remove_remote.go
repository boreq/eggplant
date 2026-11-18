package remote

import (
	"github.com/boreq/eggplant/application/accessctx"
	remotedomain "github.com/boreq/eggplant/domain/remote"
	"github.com/boreq/errors"
)

type RemoveRemote struct {
	id remotedomain.RemoteInstanceID
}

func NewRemoveRemote(id remotedomain.RemoteInstanceID) RemoveRemote {
	return RemoveRemote{
		id: id,
	}
}

func (r RemoveRemote) ID() remotedomain.RemoteInstanceID {
	return r.id
}

type RemoveRemoteHandler struct {
	transactionProvider TransactionProvider
}

func NewRemoveRemoteHandler(transactionProvider TransactionProvider) *RemoveRemoteHandler {
	return &RemoveRemoteHandler{transactionProvider: transactionProvider}
}

func (h *RemoveRemoteHandler) Execute(accessCtx accessctx.AccessContext, cmd RemoveRemote) error {
	if !accessCtx.Can(accessctx.PermissionManageRemotes) {
		return accessctx.ErrPermissionDenied
	}

	if err := h.transactionProvider.Write(func(r *TransactableRepositories) error {
		return r.RemoteInstances().Delete(cmd.ID())
	}); err != nil {
		return errors.Wrap(err, "transaction failed")
	}

	return nil
}
