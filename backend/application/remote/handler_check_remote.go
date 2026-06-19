package remote

import (
	"context"
	"time"

	remotedomain "github.com/boreq/eggplant/domain/remote"
	"github.com/boreq/errors"
)

type CheckRemote struct {
	ID remotedomain.RemoteInstanceID
}

type CheckRemoteHandler struct {
	transactionProvider TransactionProvider
	client              RemoteClient
}

func NewCheckRemoteHandler(transactionProvider TransactionProvider, client RemoteClient) *CheckRemoteHandler {
	return &CheckRemoteHandler{
		transactionProvider: transactionProvider,
		client:              client,
	}
}

func (h *CheckRemoteHandler) Execute(ctx context.Context, cmd CheckRemote) error {
	var address remotedomain.RemoteInstanceAddress
	if err := h.transactionProvider.Read(func(r *TransactableRepositories) error {
		instance, err := r.RemoteInstances.GetByID(cmd.ID)
		if err != nil {
			return errors.Wrap(err, "could not get the remote instance")
		}
		address = instance.Address()
		return nil
	}); err != nil {
		return errors.Wrap(err, "could not read the remote instance")
	}

	status := remotedomain.HealthcheckStatusAlive
	if err := h.client.Healthcheck(ctx, address); err != nil {
		status = remotedomain.HealthcheckStatusDead
	}

	return h.transactionProvider.Write(func(r *TransactableRepositories) error {
		instance, err := r.RemoteInstances.GetByID(cmd.ID)
		if err != nil {
			return errors.Wrap(err, "could not get the remote instance")
		}

		if err := instance.RecordHealthcheck(status, time.Now()); err != nil {
			return errors.Wrap(err, "could not record the healthcheck")
		}

		return r.RemoteInstances.Save(instance)
	})
}
