package remote

import (
	"context"
	"errors"

	remotedomain "github.com/boreq/eggplant/domain/remote"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")
)

// GetBy methods return ErrNotFound when no instance matches.
type RemoteInstanceRepository interface {
	Save(instance *remotedomain.RemoteInstance) error
	Delete(id remotedomain.RemoteInstanceID) error

	GetAll() ([]*remotedomain.RemoteInstance, error)
	GetByID(id remotedomain.RemoteInstanceID) (*remotedomain.RemoteInstance, error)
	GetByLocalPairingTokenHash(hash remotedomain.HashedPairingToken) (*remotedomain.RemoteInstance, error)
	GetByLocalAuthTokenHash(hash remotedomain.HashedAuthToken) (*remotedomain.RemoteInstance, error)
}

// AddEvents must run in the same write transaction as the aggregate save.
type OutboxRepository interface {
	AddEvents(source remotedomain.EventSource) error
}

type TransactionProvider interface {
	Read(handler TransactionHandler) error
	Write(handler TransactionHandler) error
}

type TransactionHandler func(repositories *TransactableRepositories) error

type TransactableRepositories struct {
	remoteInstances RemoteInstanceRepository
	outbox          OutboxRepository
}

func NewTransactableRepositories(
	remoteInstances RemoteInstanceRepository,
	outbox OutboxRepository,
) *TransactableRepositories {
	return &TransactableRepositories{
		remoteInstances: remoteInstances,
		outbox:          outbox,
	}
}

func (r TransactableRepositories) RemoteInstances() RemoteInstanceRepository {
	return r.remoteInstances
}

func (r TransactableRepositories) Outbox() OutboxRepository {
	return r.outbox
}

type RemoteClient interface {
	SendLocalAuthToken(ctx context.Context, address remotedomain.RemoteInstanceAddress, remotePairingToken remotedomain.PairingToken, localAuthToken remotedomain.AuthToken) error
	Healthcheck(ctx context.Context, address remotedomain.RemoteInstanceAddress, authToken remotedomain.AuthToken) error
}

type Remote struct {
	addRemote             *AddRemoteHandler
	removeRemote          *RemoveRemoteHandler
	setRemotePairingToken *SetRemotePairingTokenHandler
	sendLocalAuthToken    *SendLocalAuthTokenHandler
	setRemoteAuthToken    *SetRemoteAuthTokenHandler
	checkLocalAuthToken   *CheckLocalAuthTokenHandler
	listRemotes           *ListRemotesHandler
	listRemoteLibraries   *ListRemoteLibrariesHandler
	checkRemotes          *CheckRemotesHandler
	checkRemote           *CheckRemoteHandler
}

func NewRemote(
	addRemote *AddRemoteHandler,
	removeRemote *RemoveRemoteHandler,
	setRemotePairingToken *SetRemotePairingTokenHandler,
	sendLocalAuthToken *SendLocalAuthTokenHandler,
	setRemoteAuthToken *SetRemoteAuthTokenHandler,
	checkLocalAuthToken *CheckLocalAuthTokenHandler,
	listRemotes *ListRemotesHandler,
	listRemoteLibraries *ListRemoteLibrariesHandler,
	checkRemotes *CheckRemotesHandler,
	checkRemote *CheckRemoteHandler,
) *Remote {
	return &Remote{
		addRemote:             addRemote,
		removeRemote:          removeRemote,
		setRemotePairingToken: setRemotePairingToken,
		sendLocalAuthToken:    sendLocalAuthToken,
		setRemoteAuthToken:    setRemoteAuthToken,
		checkLocalAuthToken:   checkLocalAuthToken,
		listRemotes:           listRemotes,
		listRemoteLibraries:   listRemoteLibraries,
		checkRemotes:          checkRemotes,
		checkRemote:           checkRemote,
	}
}

func (r Remote) AddRemote() *AddRemoteHandler {
	return r.addRemote
}

func (r Remote) RemoveRemote() *RemoveRemoteHandler {
	return r.removeRemote
}

func (r Remote) SetRemotePairingToken() *SetRemotePairingTokenHandler {
	return r.setRemotePairingToken
}

func (r Remote) SendLocalAuthToken() *SendLocalAuthTokenHandler {
	return r.sendLocalAuthToken
}

func (r Remote) SetRemoteAuthToken() *SetRemoteAuthTokenHandler {
	return r.setRemoteAuthToken
}

func (r Remote) CheckLocalAuthToken() *CheckLocalAuthTokenHandler {
	return r.checkLocalAuthToken
}

func (r Remote) ListRemotes() *ListRemotesHandler {
	return r.listRemotes
}

func (r Remote) ListRemoteLibraries() *ListRemoteLibrariesHandler {
	return r.listRemoteLibraries
}

func (r Remote) CheckRemotes() *CheckRemotesHandler {
	return r.checkRemotes
}

func (r Remote) CheckRemote() *CheckRemoteHandler {
	return r.checkRemote
}
