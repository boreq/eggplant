package remote

import (
	"github.com/boreq/eggplant/application/accessctx"
	remotedomain "github.com/boreq/eggplant/domain/remote"
	"github.com/boreq/errors"
)

type RemoteLibrary struct {
	id   remotedomain.RemoteInstanceID
	name remotedomain.RemoteInstanceName
}

func NewRemoteLibrary(id remotedomain.RemoteInstanceID, name remotedomain.RemoteInstanceName) RemoteLibrary {
	return RemoteLibrary{
		id:   id,
		name: name,
	}
}

func (l RemoteLibrary) ID() remotedomain.RemoteInstanceID {
	return l.id
}

func (l RemoteLibrary) Name() remotedomain.RemoteInstanceName {
	return l.name
}

type ListRemoteLibrariesHandler struct {
	transactionProvider TransactionProvider
}

func NewListRemoteLibrariesHandler(transactionProvider TransactionProvider) *ListRemoteLibrariesHandler {
	return &ListRemoteLibrariesHandler{transactionProvider: transactionProvider}
}

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
			name, err := instance.Name()
			if err != nil {
				return errors.Wrap(err, "could not get the name")
			}
			libraries = append(libraries, NewRemoteLibrary(instance.Id(), name))
		}
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}

	return libraries, nil
}
