package auth_test

import (
	"testing"

	app "github.com/boreq/eggplant/application/auth"
	authdomain "github.com/boreq/eggplant/domain/auth"
	"github.com/boreq/eggplant/internal/fixture"
	"github.com/stretchr/testify/require"
)

func TestSessionTokenPutAndGet(t *testing.T) {
	tp, cleanup := fixture.AuthTransactionProvider(t)
	defer cleanup()

	token, err := authdomain.NewAccessTokenFromString("token")
	require.NoError(t, err)
	username, err := authdomain.NewUsernameFromString("username")
	require.NoError(t, err)

	err = tp.Write(func(r *app.TransactableRepositories) error {
		return r.SessionTokens.Put(token, username)
	})
	require.NoError(t, err)

	err = tp.Read(func(r *app.TransactableRepositories) error {
		got, err := r.SessionTokens.Get(token)
		require.NoError(t, err)
		require.Equal(t, username, got)
		return nil
	})
	require.NoError(t, err)
}

func TestSessionTokenGetNotFound(t *testing.T) {
	tp, cleanup := fixture.AuthTransactionProvider(t)
	defer cleanup()

	token, err := authdomain.NewAccessTokenFromString("token")
	require.NoError(t, err)

	err = tp.Read(func(r *app.TransactableRepositories) error {
		_, err := r.SessionTokens.Get(token)
		require.ErrorIs(t, err, app.ErrNotFound)
		return nil
	})
	require.NoError(t, err)
}

func TestSessionTokenPutOverwrites(t *testing.T) {
	tp, cleanup := fixture.AuthTransactionProvider(t)
	defer cleanup()

	token, err := authdomain.NewAccessTokenFromString("token")
	require.NoError(t, err)
	username1, err := authdomain.NewUsernameFromString("user1")
	require.NoError(t, err)
	username2, err := authdomain.NewUsernameFromString("user2")
	require.NoError(t, err)

	err = tp.Write(func(r *app.TransactableRepositories) error {
		if err := r.SessionTokens.Put(token, username1); err != nil {
			return err
		}
		return r.SessionTokens.Put(token, username2)
	})
	require.NoError(t, err)

	err = tp.Read(func(r *app.TransactableRepositories) error {
		got, err := r.SessionTokens.Get(token)
		require.NoError(t, err)
		require.Equal(t, username2, got)
		return nil
	})
	require.NoError(t, err)
}

func TestSessionTokenRemove(t *testing.T) {
	tp, cleanup := fixture.AuthTransactionProvider(t)
	defer cleanup()

	token, err := authdomain.NewAccessTokenFromString("token")
	require.NoError(t, err)
	username, err := authdomain.NewUsernameFromString("username")
	require.NoError(t, err)

	err = tp.Write(func(r *app.TransactableRepositories) error {
		return r.SessionTokens.Put(token, username)
	})
	require.NoError(t, err)

	err = tp.Write(func(r *app.TransactableRepositories) error {
		return r.SessionTokens.Remove(token)
	})
	require.NoError(t, err)

	err = tp.Read(func(r *app.TransactableRepositories) error {
		_, err := r.SessionTokens.Get(token)
		require.ErrorIs(t, err, app.ErrNotFound)
		return nil
	})
	require.NoError(t, err)
}

func TestSessionTokenRemoveMissingIsNoop(t *testing.T) {
	tp, cleanup := fixture.AuthTransactionProvider(t)
	defer cleanup()

	token, err := authdomain.NewAccessTokenFromString("token")
	require.NoError(t, err)

	err = tp.Write(func(r *app.TransactableRepositories) error {
		return r.SessionTokens.Remove(token)
	})
	require.NoError(t, err)
}

func TestSessionTokensIndependent(t *testing.T) {
	tp, cleanup := fixture.AuthTransactionProvider(t)
	defer cleanup()

	tokenA, err := authdomain.NewAccessTokenFromString("a")
	require.NoError(t, err)
	tokenB, err := authdomain.NewAccessTokenFromString("b")
	require.NoError(t, err)
	usernameA, err := authdomain.NewUsernameFromString("ua")
	require.NoError(t, err)
	usernameB, err := authdomain.NewUsernameFromString("ub")
	require.NoError(t, err)

	err = tp.Write(func(r *app.TransactableRepositories) error {
		if err := r.SessionTokens.Put(tokenA, usernameA); err != nil {
			return err
		}
		return r.SessionTokens.Put(tokenB, usernameB)
	})
	require.NoError(t, err)

	err = tp.Write(func(r *app.TransactableRepositories) error {
		return r.SessionTokens.Remove(tokenA)
	})
	require.NoError(t, err)

	err = tp.Read(func(r *app.TransactableRepositories) error {
		_, err := r.SessionTokens.Get(tokenA)
		require.ErrorIs(t, err, app.ErrNotFound)

		got, err := r.SessionTokens.Get(tokenB)
		require.NoError(t, err)
		require.Equal(t, usernameB, got)
		return nil
	})
	require.NoError(t, err)
}
