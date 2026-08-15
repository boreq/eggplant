package remote

import (
	"github.com/boreq/eggplant/application/accessctx"
	remotedomain "github.com/boreq/eggplant/domain/remote"
	"github.com/boreq/errors"
)

type RemoteLibrary struct {
	id      remotedomain.RemoteInstanceID
	address remotedomain.RemoteInstanceAddress
}

func NewRemoteLibrary(id remotedomain.RemoteInstanceID, address remotedomain.RemoteInstanceAddress) RemoteLibrary {
	return RemoteLibrary{
		id:      id,
		address: address,
	}
}

func (l RemoteLibrary) ID() remotedomain.RemoteInstanceID {
	return l.id
}

func (l RemoteLibrary) Address() remotedomain.RemoteInstanceAddress {
	return l.address
}

type ListRemoteLibrariesHandler struct {
	transactionProvider TransactionProvider
}

func NewListRemoteLibrariesHandler(transactionProvider TransactionProvider) *ListRemoteLibrariesHandler {
	return &ListRemoteLibrariesHandler{transactionProvider: transactionProvider}
}

// Execute lists the instances the libraries of which can be browsed, so the
// ones which gave us an auth token during pairing.
func (h *ListRemoteLibrariesHandler) Execute(accessCtx accessctx.AccessContext) ([]RemoteLibrary, error) {
	if !accessCtx.Can(accessctx.PermissionSeeRemoteLibraryContent) {
		return nil, accessctx.ErrPermissionDenied
	}

	var libraries []RemoteLibrary

	if err := h.transactionProvider.Read(func(r *TransactableRepositories) error {
		instances, err := r.RemoteInstances.GetAll()
		if err != nil {
			return errors.Wrap(err, "could not get the remote instances")
		}
		for _, instance := range instances {
			if !instance.CanBeQueried() {
				continue
			}
			libraries = append(libraries, NewRemoteLibrary(instance.Id(), instance.Address()))
		}
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}

	return libraries, nil
}
