package remote

import (
	"context"
	"time"

	remotedomain "github.com/boreq/eggplant/domain/remote"
	"github.com/boreq/eggplant/internal/logging"
	"github.com/boreq/errors"
)

type CheckRemotesHandler struct {
	transactionProvider TransactionProvider
	client              RemoteClient
	log                 logging.Logger
}

func NewCheckRemotesHandler(transactionProvider TransactionProvider, client RemoteClient) *CheckRemotesHandler {
	return &CheckRemotesHandler{
		transactionProvider: transactionProvider,
		client:              client,
		log:                 logging.New("application/remote.CheckRemotesHandler"),
	}
}

func (h *CheckRemotesHandler) Execute(ctx context.Context) error {
	var instances []*remotedomain.RemoteInstance
	if err := h.transactionProvider.Read(func(r *TransactableRepositories) error {
		all, err := r.RemoteInstances.GetAll()
		if err != nil {
			return errors.Wrap(err, "could not get the remote instances")
		}
		instances = all
		return nil
	}); err != nil {
		return errors.Wrap(err, "could not list the remote instances")
	}

	for _, instance := range instances {
		status := remotedomain.HealthcheckStatusAlive
		if err := h.client.Healthcheck(ctx, instance.Address()); err != nil {
			h.log.Debug("healthcheck failed", "id", instance.Id().String(), "err", err)
			status = remotedomain.HealthcheckStatusDead
		}

		if err := h.record(instance.Id(), status); err != nil {
			h.log.Error("could not record the healthcheck", "id", instance.Id().String(), "err", err)
		}
	}

	return nil
}

func (h *CheckRemotesHandler) record(id remotedomain.RemoteInstanceID, status remotedomain.HealthcheckStatus) error {
	return h.transactionProvider.Write(func(r *TransactableRepositories) error {
		instance, err := r.RemoteInstances.GetByID(id)
		if err != nil {
			return errors.Wrap(err, "could not get the remote instance")
		}

		if err := instance.RecordHealthcheck(status, time.Now()); err != nil {
			return errors.Wrap(err, "could not record the healthcheck")
		}

		return r.RemoteInstances.Save(instance)
	})
}
