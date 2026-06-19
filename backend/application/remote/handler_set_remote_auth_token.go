package remote

import (
	remotedomain "github.com/boreq/eggplant/domain/remote"
	"github.com/boreq/errors"
)

type SetRemoteAuthToken struct {
	LocalPairingToken remotedomain.PairingToken
	RemoteAuthToken   remotedomain.AuthToken
}

type SetRemoteAuthTokenHandler struct {
	transactionProvider TransactionProvider
}

func NewSetRemoteAuthTokenHandler(transactionProvider TransactionProvider) *SetRemoteAuthTokenHandler {
	return &SetRemoteAuthTokenHandler{transactionProvider: transactionProvider}
}

func (h *SetRemoteAuthTokenHandler) Execute(cmd SetRemoteAuthToken) error {
	if err := h.transactionProvider.Write(func(r *TransactableRepositories) error {
		instance, err := r.RemoteInstances.GetByLocalPairingTokenHash(cmd.LocalPairingToken.Hash())
		if err != nil {
			return errors.Wrap(err, "could not get the remote instance")
		}

		if err := instance.SetRemoteAuthToken(cmd.LocalPairingToken, cmd.RemoteAuthToken); err != nil {
			return errors.Wrap(err, "could not store the remote auth token")
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
