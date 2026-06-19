package fixture

import (
	"testing"

	"github.com/boreq/eggplant/adapters/remote"
	appremote "github.com/boreq/eggplant/application/remote"
	"github.com/boreq/eggplant/internal/wire"
	bolt "go.etcd.io/bbolt"
)

type remoteRepositoriesProvider struct{}

func RemoteTransactionProvider(t *testing.T) (appremote.TransactionProvider, *bolt.DB, CleanupFunc) {
	db, cleanup := Bolt(t)
	return remote.NewRemoteTransactionProvider(db, &remoteRepositoriesProvider{}), db, cleanup
}

func (p *remoteRepositoriesProvider) Provide(tx *bolt.Tx) (*appremote.TransactableRepositories, error) {
	return wire.BuildTransactableRemoteRepositories(tx)
}
