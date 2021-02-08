package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/boreq/eggplant/adapters/auth"
	app "github.com/boreq/eggplant/application/auth"
	"github.com/boreq/eggplant/internal/fixture"
	"github.com/boreq/eggplant/internal/wire"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
)

func TestLastSeenUpdater(t *testing.T) {
	db, cleanup := fixture.Bolt(t)
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repositoriesProvider := newAuthRepositoriesProvider()

	transactionProvider := auth.NewAuthTransactionProvider(
		db,
		repositoriesProvider,
	)

	u, err := auth.NewLastSeenUpdater(transactionProvider)
	require.NoError(t, err)

	username := "username"

	session1 := app.Session{
		Token:    "a",
		LastSeen: time.Now().Add(-10 * time.Second),
	}

	session2 := app.Session{
		Token:    "b",
		LastSeen: time.Now().Add(-10 * time.Second),
	}

	err = transactionProvider.Write(func(adapters *app.TransactableRepositories) error {
		return adapters.Users.Put(app.User{
			Username: username,
			Sessions: []app.Session{
				session1,
				session2,
			},
		})
	})
	require.NoError(t, err)

	newValue := time.Now()
	u.Update(username, session1.Token, newValue)

	go func() {
		u.Run(ctx, time.Second)
	}()
	<-time.After(2 * time.Second)

	err = transactionProvider.Read(func(adapters *app.TransactableRepositories) error {
		u, err := adapters.Users.Get(username)
		require.NoError(t, err)

		require.Len(t, u.Sessions, 2)

		require.Equal(t, session1.Token, u.Sessions[0].Token)
		require.True(t, u.Sessions[0].LastSeen.Equal(newValue))

		require.Equal(t, session2.Token, u.Sessions[1].Token)
		require.True(t, u.Sessions[1].LastSeen.Equal(session2.LastSeen))

		return nil
	})
	require.NoError(t, err)
}

type authRepositoriesProvider struct {
}

func newAuthRepositoriesProvider() *authRepositoriesProvider {
	return &authRepositoriesProvider{}
}

func (p *authRepositoriesProvider) Provide(tx *bolt.Tx) (*app.TransactableRepositories, error) {
	return wire.BuildTransactableAuthRepositories(tx)
}
