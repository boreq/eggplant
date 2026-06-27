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
	RemoteInstances RemoteInstanceRepository
	Outbox          OutboxRepository
}

type RemoteClient interface {
	SendLocalAuthToken(ctx context.Context, address remotedomain.RemoteInstanceAddress, remotePairingToken remotedomain.PairingToken, localAuthToken remotedomain.AuthToken) error
	Healthcheck(ctx context.Context, address remotedomain.RemoteInstanceAddress, authToken remotedomain.AuthToken) error
}

type Remote struct {
	AddRemote             *AddRemoteHandler
	SetRemotePairingToken *SetRemotePairingTokenHandler
	SendLocalAuthToken    *SendLocalAuthTokenHandler
	SetRemoteAuthToken    *SetRemoteAuthTokenHandler
	CheckLocalAuthToken   *CheckLocalAuthTokenHandler
	ListRemotes           *ListRemotesHandler
	CheckRemotes          *CheckRemotesHandler
	CheckRemote           *CheckRemoteHandler
}
