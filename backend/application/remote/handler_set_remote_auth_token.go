package remote

import (
	remotedomain "github.com/boreq/eggplant/domain/remote"
	"github.com/boreq/errors"
)

type SetRemoteAuthToken struct {
	localPairingToken remotedomain.PairingToken
	remoteAuthToken   remotedomain.AuthToken
}

func NewSetRemoteAuthToken(localPairingToken remotedomain.PairingToken, remoteAuthToken remotedomain.AuthToken) SetRemoteAuthToken {
	return SetRemoteAuthToken{
		localPairingToken: localPairingToken,
		remoteAuthToken:   remoteAuthToken,
	}
}

func (s SetRemoteAuthToken) LocalPairingToken() remotedomain.PairingToken {
	return s.localPairingToken
}

func (s SetRemoteAuthToken) RemoteAuthToken() remotedomain.AuthToken {
	return s.remoteAuthToken
}

type SetRemoteAuthTokenHandler struct {
	transactionProvider TransactionProvider
}

func NewSetRemoteAuthTokenHandler(transactionProvider TransactionProvider) *SetRemoteAuthTokenHandler {
	return &SetRemoteAuthTokenHandler{transactionProvider: transactionProvider}
}

func (h *SetRemoteAuthTokenHandler) Execute(cmd SetRemoteAuthToken) error {
	if err := h.transactionProvider.Write(func(r *TransactableRepositories) error {
		instance, err := r.RemoteInstances.GetByLocalPairingTokenHash(cmd.LocalPairingToken().Hash())
		if err != nil {
			return errors.Wrap(err, "could not get the remote instance")
		}

		if err := instance.SetRemoteAuthToken(cmd.LocalPairingToken(), cmd.RemoteAuthToken()); err != nil {
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
