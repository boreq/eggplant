package remote

import (
	"github.com/boreq/eggplant/application/accessctx"
	remotedomain "github.com/boreq/eggplant/domain/remote"
	"github.com/boreq/errors"
)

type CheckLocalAuthToken struct {
	Token remotedomain.AuthToken
}

type CheckLocalAuthTokenHandler struct {
	transactionProvider TransactionProvider
}

func NewCheckLocalAuthTokenHandler(transactionProvider TransactionProvider) *CheckLocalAuthTokenHandler {
	return &CheckLocalAuthTokenHandler{transactionProvider: transactionProvider}
}

func (h *CheckLocalAuthTokenHandler) Execute(cmd CheckLocalAuthToken) (accessctx.RemoteInstanceAccessContext, error) {
	var remoteCtx accessctx.RemoteInstanceAccessContext

	if err := h.transactionProvider.Read(func(r *TransactableRepositories) error {
		instance, err := r.RemoteInstances.GetByLocalAuthTokenHash(cmd.Token.Hash())
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return ErrUnauthorized
			}
			return errors.Wrap(err, "could not get the remote instance")
		}
		remoteCtx = accessctx.NewRemoteInstanceAccessContext(instance.Id())
		return nil
	}); err != nil {
		return accessctx.RemoteInstanceAccessContext{}, errors.Wrap(err, "transaction failed")
	}

	return remoteCtx, nil
}
