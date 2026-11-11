package remote

import (
	"github.com/boreq/eggplant/application/accessctx"
	remotedomain "github.com/boreq/eggplant/domain/remote"
	"github.com/boreq/errors"
)

type AddRemote struct {
	address remotedomain.RemoteInstanceAddress
}

func NewAddRemote(address remotedomain.RemoteInstanceAddress) AddRemote {
	return AddRemote{
		address: address,
	}
}

func (a AddRemote) Address() remotedomain.RemoteInstanceAddress {
	return a.address
}

type AddRemoteResult struct {
	id                remotedomain.RemoteInstanceID
	localPairingToken remotedomain.PairingToken
}

func NewAddRemoteResult(id remotedomain.RemoteInstanceID, localPairingToken remotedomain.PairingToken) AddRemoteResult {
	return AddRemoteResult{
		id:                id,
		localPairingToken: localPairingToken,
	}
}

func (r AddRemoteResult) ID() remotedomain.RemoteInstanceID {
	return r.id
}

func (r AddRemoteResult) LocalPairingToken() remotedomain.PairingToken {
	return r.localPairingToken
}

type AddRemoteHandler struct {
	transactionProvider TransactionProvider
}

func NewAddRemoteHandler(transactionProvider TransactionProvider) *AddRemoteHandler {
	return &AddRemoteHandler{transactionProvider: transactionProvider}
}

func (h *AddRemoteHandler) Execute(accessCtx accessctx.AccessContext, cmd AddRemote) (AddRemoteResult, error) {
	if !accessCtx.Can(accessctx.PermissionManageRemotes) {
		return AddRemoteResult{}, accessctx.ErrPermissionDenied
	}

	id, err := remotedomain.NewRemoteInstanceID()
	if err != nil {
		return AddRemoteResult{}, errors.Wrap(err, "could not generate the id")
	}

	localPairingToken, err := remotedomain.NewPairingToken()
	if err != nil {
		return AddRemoteResult{}, errors.Wrap(err, "could not generate the local pairing token")
	}

	instance := remotedomain.NewRemoteInstance(id, cmd.Address(), localPairingToken.Hash())

	if err := h.transactionProvider.Write(func(r *TransactableRepositories) error {
		return r.RemoteInstances().Save(instance)
	}); err != nil {
		return AddRemoteResult{}, errors.Wrap(err, "transaction failed")
	}

	return NewAddRemoteResult(id, localPairingToken), nil
}
