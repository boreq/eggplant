package auth_test

import (
	"context"
	"testing"
	"time"

	app "github.com/boreq/eggplant/application/auth"
	authdomain "github.com/boreq/eggplant/domain/auth"
	"github.com/boreq/eggplant/internal/fixture"
	"github.com/stretchr/testify/require"
)

type lastSeenUpdaterStub struct {
	updates []app.LastSeenUpdate
}

func (s *lastSeenUpdaterStub) Update(username authdomain.Username, token authdomain.AccessToken, t time.Time) {
}

func (s *lastSeenUpdaterStub) PopUpdates() []app.LastSeenUpdate {
	return s.updates
}

func TestPersistLastSeenHandler_PersistsBufferedUpdates(t *testing.T) {
	transactionProvider, cleanup := fixture.AuthTransactionProvider(t)
	defer cleanup()

	username, err := authdomain.NewUsernameFromString("username")
	require.NoError(t, err)

	tokenA, err := authdomain.NewAccessTokenFromString("a")
	require.NoError(t, err)
	tokenB, err := authdomain.NewAccessTokenFromString("b")
	require.NoError(t, err)

	session2Time := time.Now().Add(-10 * time.Second)
	session1 := authdomain.NewSession(tokenA, session2Time)
	session2 := authdomain.NewSession(tokenB, session2Time)

	password, err := authdomain.NewPasswordHash([]byte("hash"))
	require.NoError(t, err)

	now := time.Now()
	user := authdomain.NewUserFromDatabase(
		username,
		password,
		false,
		&now,
		&now,
		[]authdomain.Session{session1, session2},
	)

	err = transactionProvider.Write(func(adapters *app.TransactableRepositories) error {
		return adapters.Users.Put(user)
	})
	require.NoError(t, err)

	newValue := time.Now()
	stub := &lastSeenUpdaterStub{
		updates: []app.LastSeenUpdate{
			{
				Username: username,
				LastSeen: newValue,
				Sessions: map[authdomain.AccessToken]time.Time{tokenA: newValue},
			},
		},
	}

	handler := app.NewPersistLastSeenHandler(transactionProvider, stub)
	require.NoError(t, handler.Execute(context.Background()))

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

func TestPersistLastSeenHandler_SkipsUnknownUser(t *testing.T) {
	transactionProvider, cleanup := fixture.AuthTransactionProvider(t)
	defer cleanup()

	username, err := authdomain.NewUsernameFromString("ghost")
	require.NoError(t, err)

	token, err := authdomain.NewAccessTokenFromString("a")
	require.NoError(t, err)

	stub := &lastSeenUpdaterStub{
		updates: []app.LastSeenUpdate{
			{
				Username: username,
				LastSeen: time.Now(),
				Sessions: map[authdomain.AccessToken]time.Time{token: time.Now()},
			},
		},
	}

	handler := app.NewPersistLastSeenHandler(transactionProvider, stub)
	require.NoError(t, handler.Execute(context.Background()))
}
