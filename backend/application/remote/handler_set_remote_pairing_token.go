package remote

import (
	"github.com/boreq/eggplant/application/accessctx"
	remotedomain "github.com/boreq/eggplant/domain/remote"
	"github.com/boreq/errors"
)

type SetRemotePairingToken struct {
	ID                 remotedomain.RemoteInstanceID
	RemotePairingToken remotedomain.PairingToken
}

type SetRemotePairingTokenHandler struct {
	transactionProvider TransactionProvider
}

func NewSetRemotePairingTokenHandler(transactionProvider TransactionProvider) *SetRemotePairingTokenHandler {
	return &SetRemotePairingTokenHandler{transactionProvider: transactionProvider}
}

func (h *SetRemotePairingTokenHandler) Execute(accessCtx accessctx.AccessContext, cmd SetRemotePairingToken) error {
	if !accessCtx.Can(accessctx.PermissionManageRemotes) {
		return accessctx.ErrPermissionDenied
	}

	if err := h.transactionProvider.Write(func(r *TransactableRepositories) error {
		instance, err := r.RemoteInstances.GetByID(cmd.ID)
		if err != nil {
			return errors.Wrap(err, "could not get the remote instance")
		}

		if err := instance.SetRemotePairingToken(cmd.RemotePairingToken); err != nil {
			return errors.Wrap(err, "could not set the remote pairing token")
		}

		if err := r.RemoteInstances.Save(instance); err != nil {
			return errors.Wrap(err, "could not save the remote instance")
		}

		if err := r.Outbox.AddEvents(instance); err != nil {
			return errors.Wrap(err, "could not add the events to the outbox")
		}

		return nil
	}); err != nil {
		return errors.Wrap(err, "transaction failed")
	}

	return nil
}
