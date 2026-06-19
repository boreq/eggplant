package remote

import (
	"github.com/boreq/eggplant/application/remote"
	"github.com/boreq/errors"
	bolt "go.etcd.io/bbolt"
)

type RemoteRepositoriesProvider interface {
	Provide(tx *bolt.Tx) (*remote.TransactableRepositories, error)
}

type RemoteTransactionProvider struct {
	db                   *bolt.DB
	repositoriesProvider RemoteRepositoriesProvider
}

func NewRemoteTransactionProvider(
	db *bolt.DB,
	repositoriesProvider RemoteRepositoriesProvider,
) *RemoteTransactionProvider {
	return &RemoteTransactionProvider{
		db:                   db,
		repositoriesProvider: repositoriesProvider,
	}
}

func (p *RemoteTransactionProvider) Read(handler remote.TransactionHandler) error {
	return p.db.View(func(tx *bolt.Tx) error {
		repositories, err := p.repositoriesProvider.Provide(tx)
		if err != nil {
			return errors.Wrap(err, "could not provide the repositories")
		}
		return handler(repositories)
	})
}

func (p *RemoteTransactionProvider) Write(handler remote.TransactionHandler) error {
	return p.db.Batch(func(tx *bolt.Tx) error {
		repositories, err := p.repositoriesProvider.Provide(tx)
		if err != nil {
			return errors.Wrap(err, "could not provide the repositories")
		}
		return handler(repositories)
	})
}
