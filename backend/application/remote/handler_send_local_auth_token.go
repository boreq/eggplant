package remote

import (
	"context"

	remotedomain "github.com/boreq/eggplant/domain/remote"
	"github.com/boreq/errors"
)

type SendLocalAuthToken struct {
	RemoteInstanceID remotedomain.RemoteInstanceID
}

type SendLocalAuthTokenHandler struct {
	transactionProvider TransactionProvider
	client              RemoteClient
}

func NewSendLocalAuthTokenHandler(
	transactionProvider TransactionProvider,
	client RemoteClient,
) *SendLocalAuthTokenHandler {
	return &SendLocalAuthTokenHandler{
		transactionProvider: transactionProvider,
		client:              client,
	}
}

func (h *SendLocalAuthTokenHandler) Execute(ctx context.Context, cmd SendLocalAuthToken) error {
	var address remotedomain.RemoteInstanceAddress
	var remotePairingToken remotedomain.PairingToken
	var localAuthToken remotedomain.AuthToken

	if err := h.transactionProvider.Write(func(r *TransactableRepositories) error {
		instance, err := r.RemoteInstances.GetByID(cmd.RemoteInstanceID)
		if err != nil {
			return errors.Wrap(err, "could not get the remote instance")
		}

		pt, ok := instance.RemotePairingToken()
		if !ok {
			return errors.New("remote pairing token is not set")
		}

		token, err := instance.IssueLocalAuthToken()
		if err != nil {
			return errors.Wrap(err, "could not issue the local auth token")
		}

		remotePairingToken = pt
		localAuthToken = token
		address = instance.Address()

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

	if err := h.client.SendLocalAuthToken(ctx, address, remotePairingToken, localAuthToken); err != nil {
		return errors.Wrap(err, "could not deliver the credential to the peer")
	}

	return nil
}
