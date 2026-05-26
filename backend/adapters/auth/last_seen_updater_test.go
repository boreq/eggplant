package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/boreq/eggplant/adapters/auth"
	app "github.com/boreq/eggplant/application/auth"
	authdomain "github.com/boreq/eggplant/domain/auth"
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

	username, err := authdomain.NewUsernameFromString("username")
	require.NoError(t, err)

	tokenA, err := authdomain.NewAccessTokenFromString("a")
	require.NoError(t, err)
	tokenB, err := authdomain.NewAccessTokenFromString("b")
	require.NoError(t, err)

	session1Time := time.Now().Add(-10 * time.Second)
	session2Time := time.Now().Add(-10 * time.Second)
	session1 := authdomain.NewSession(tokenA, session1Time)
	session2 := authdomain.NewSession(tokenB, session2Time)

	password, err := authdomain.NewPasswordHash([]byte("hash"))
	require.NoError(t, err)

	user := authdomain.NewUser(
		username,
		password,
		false,
		time.Now(),
		time.Now(),
		[]authdomain.Session{session1, session2},
	)

	err = transactionProvider.Write(func(adapters *app.TransactableRepositories) error {
		return adapters.Users.Put(user)
	})
	require.NoError(t, err)

	newValue := time.Now()
	u.Update(username, tokenA, newValue)

	go func() {
		u.Run(ctx, time.Second)
	}()
	<-time.After(2 * time.Second)

	err = transactionProvider.Read(func(adapters *app.TransactableRepositories) error {
		u, err := adapters.Users.Get(username)
		require.NoError(t, err)

		require.Len(t, u.Sessions(), 2)

		require.Equal(t, tokenA, u.Sessions()[0].Token())
		require.True(t, u.Sessions()[0].LastSeen().Equal(newValue))

		require.Equal(t, tokenB, u.Sessions()[1].Token())
		require.True(t, u.Sessions()[1].LastSeen().Equal(session2Time))

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
