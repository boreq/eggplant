package fixture

import (
	"testing"

	"github.com/boreq/eggplant/adapters/auth"
	appauth "github.com/boreq/eggplant/application/auth"
	"github.com/boreq/eggplant/internal/wire"
	bolt "go.etcd.io/bbolt"
)

type authRepositoriesProvider struct{}

func AuthTransactionProvider(t *testing.T) (appauth.TransactionProvider, CleanupFunc) {
	db, cleanup := Bolt(t)
	return auth.NewAuthTransactionProvider(db, &authRepositoriesProvider{}), cleanup
}

func (p *authRepositoriesProvider) Provide(tx *bolt.Tx) (*appauth.TransactableRepositories, error) {
	return wire.BuildTransactableAuthRepositories(tx)
}
