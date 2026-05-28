package auth_test

import (
	"testing"
	"time"

	app "github.com/boreq/eggplant/application/auth"
	authdomain "github.com/boreq/eggplant/domain/auth"
	"github.com/boreq/eggplant/internal/fixture"
	"github.com/boreq/errors"
	"github.com/stretchr/testify/require"
)

func TestRemoveOldSessionsWhenSavingUsers(t *testing.T) {
	tp, cleanup := fixture.AuthTransactionProvider(t)
	defer cleanup()

	tokenOld, err := authdomain.NewAccessTokenFromString("a")
	require.NoError(t, err)
	tokenRecent, err := authdomain.NewAccessTokenFromString("b")
	require.NoError(t, err)

	session1 := authdomain.NewSession(tokenOld, time.Now().Add(-2*365*24*time.Hour))
	session2 := authdomain.NewSession(tokenRecent, time.Now().Add(-10*24*time.Hour))

	username, err := authdomain.NewUsernameFromString("username")
	require.NoError(t, err)
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

	err = tp.Write(func(r *app.TransactableRepositories) error {
		return r.Users.Put(user)
	})
	require.NoError(t, err)

	err = tp.Read(func(r *app.TransactableRepositories) error {
		u, err := r.Users.Get(username)
		if err != nil {
			return errors.Wrap(err, "get failed")
		}

		require.Len(t, u.Sessions(), 1)
		require.Equal(t, tokenRecent, u.Sessions()[0].Token())

		return nil
	})
	require.NoError(t, err)
}

func TestSaveAndLoadUser(t *testing.T) {
	tp, cleanup := fixture.AuthTransactionProvider(t)
	defer cleanup()

	username, err := authdomain.NewUsernameFromString("username")
	require.NoError(t, err)
	password, err := authdomain.NewPasswordHash([]byte("hash"))
	require.NoError(t, err)

	token, err := authdomain.NewAccessTokenFromString("token")
	require.NoError(t, err)

	created := time.Now().Add(-24 * time.Hour)
	lastSeen := time.Now().Add(-1 * time.Hour)
	sessionLastSeen := time.Now().Add(-30 * time.Minute)

	session := authdomain.NewSession(token, sessionLastSeen)

	user := authdomain.NewUserFromDatabase(
		username,
		password,
		true,
		&created,
		&lastSeen,
		[]authdomain.Session{session},
	)

	err = tp.Write(func(r *app.TransactableRepositories) error {
		return r.Users.Put(user)
	})
	require.NoError(t, err)

	err = tp.Read(func(r *app.TransactableRepositories) error {
		u, err := r.Users.Get(username)
		if err != nil {
			return errors.Wrap(err, "get failed")
		}

		require.Equal(t, username, u.Username())
		require.Equal(t, password.Bytes(), u.Password().Bytes())
		require.True(t, u.Administrator())
		require.NotNil(t, u.Created())
		require.True(t, created.Equal(*u.Created()))
		require.NotNil(t, u.LastSeen())
		require.True(t, lastSeen.Equal(*u.LastSeen()))
		require.Len(t, u.Sessions(), 1)
		require.Equal(t, token, u.Sessions()[0].Token())
		require.True(t, sessionLastSeen.Equal(u.Sessions()[0].LastSeen()))

		return nil
	})
	require.NoError(t, err)
}

func TestGetByToken(t *testing.T) {
	tp, cleanup := fixture.AuthTransactionProvider(t)
	defer cleanup()

	username, err := authdomain.NewUsernameFromString("username")
	require.NoError(t, err)
	password, err := authdomain.NewPasswordHash([]byte("hash"))
	require.NoError(t, err)
	token, err := authdomain.NewAccessTokenFromString("token")
	require.NoError(t, err)

	now := time.Now()
	user := authdomain.NewUserFromDatabase(
		username, password, false, &now, &now,
		[]authdomain.Session{authdomain.NewSession(token, now)},
	)

	err = tp.Write(func(r *app.TransactableRepositories) error {
		return r.Users.Put(user)
	})
	require.NoError(t, err)

	err = tp.Read(func(r *app.TransactableRepositories) error {
		u, err := r.Users.GetByToken(token)
		require.NoError(t, err)
		require.Equal(t, username, u.Username())
		return nil
	})
	require.NoError(t, err)
}

func TestGetByTokenNotFound(t *testing.T) {
	tp, cleanup := fixture.AuthTransactionProvider(t)
	defer cleanup()

	token, err := authdomain.NewAccessTokenFromString("missing")
	require.NoError(t, err)

	err = tp.Read(func(r *app.TransactableRepositories) error {
		_, err := r.Users.GetByToken(token)
		require.ErrorIs(t, err, app.ErrNotFound)
		return nil
	})
	require.NoError(t, err)
}

func TestPutIndexesNewSessions(t *testing.T) {
	tp, cleanup := fixture.AuthTransactionProvider(t)
	defer cleanup()

	username, err := authdomain.NewUsernameFromString("username")
	require.NoError(t, err)
	password, err := authdomain.NewPasswordHash([]byte("hash"))
	require.NoError(t, err)
	tokenA, err := authdomain.NewAccessTokenFromString("a")
	require.NoError(t, err)
	tokenB, err := authdomain.NewAccessTokenFromString("b")
	require.NoError(t, err)

	now := time.Now()
	user := authdomain.NewUserFromDatabase(
		username, password, false, &now, &now,
		[]authdomain.Session{
			authdomain.NewSession(tokenA, now),
			authdomain.NewSession(tokenB, now),
		},
	)

	err = tp.Write(func(r *app.TransactableRepositories) error {
		return r.Users.Put(user)
	})
	require.NoError(t, err)

	err = tp.Read(func(r *app.TransactableRepositories) error {
		u, err := r.Users.GetByToken(tokenA)
		require.NoError(t, err)
		require.Equal(t, username, u.Username())

		u, err = r.Users.GetByToken(tokenB)
		require.NoError(t, err)
		require.Equal(t, username, u.Username())
		return nil
	})
	require.NoError(t, err)
}

func TestPutRemovesStaleSessionIndex(t *testing.T) {
	tp, cleanup := fixture.AuthTransactionProvider(t)
	defer cleanup()

	username, err := authdomain.NewUsernameFromString("username")
	require.NoError(t, err)
	password, err := authdomain.NewPasswordHash([]byte("hash"))
	require.NoError(t, err)
	tokenA, err := authdomain.NewAccessTokenFromString("a")
	require.NoError(t, err)
	tokenB, err := authdomain.NewAccessTokenFromString("b")
	require.NoError(t, err)

	now := time.Now()

	err = tp.Write(func(r *app.TransactableRepositories) error {
		return r.Users.Put(authdomain.NewUserFromDatabase(
			username, password, false, &now, &now,
			[]authdomain.Session{
				authdomain.NewSession(tokenA, now),
				authdomain.NewSession(tokenB, now),
			},
		))
	})
	require.NoError(t, err)

	err = tp.Write(func(r *app.TransactableRepositories) error {
		return r.Users.Put(authdomain.NewUserFromDatabase(
			username, password, false, &now, &now,
			[]authdomain.Session{authdomain.NewSession(tokenB, now)},
		))
	})
	require.NoError(t, err)

	err = tp.Read(func(r *app.TransactableRepositories) error {
		_, err := r.Users.GetByToken(tokenA)
		require.ErrorIs(t, err, app.ErrNotFound)

		u, err := r.Users.GetByToken(tokenB)
		require.NoError(t, err)
		require.Equal(t, username, u.Username())
		return nil
	})
	require.NoError(t, err)
}

func TestPutRemovesExpiredSessionIndex(t *testing.T) {
	tp, cleanup := fixture.AuthTransactionProvider(t)
	defer cleanup()

	username, err := authdomain.NewUsernameFromString("username")
	require.NoError(t, err)
	password, err := authdomain.NewPasswordHash([]byte("hash"))
	require.NoError(t, err)
	tokenOld, err := authdomain.NewAccessTokenFromString("old")
	require.NoError(t, err)
	tokenRecent, err := authdomain.NewAccessTokenFromString("recent")
	require.NoError(t, err)

	now := time.Now()
	user := authdomain.NewUserFromDatabase(
		username, password, false, &now, &now,
		[]authdomain.Session{
			authdomain.NewSession(tokenOld, now.Add(-2*365*24*time.Hour)),
			authdomain.NewSession(tokenRecent, now),
		},
	)

	err = tp.Write(func(r *app.TransactableRepositories) error {
		return r.Users.Put(user)
	})
	require.NoError(t, err)

	err = tp.Read(func(r *app.TransactableRepositories) error {
		_, err := r.Users.GetByToken(tokenOld)
		require.ErrorIs(t, err, app.ErrNotFound)

		u, err := r.Users.GetByToken(tokenRecent)
		require.NoError(t, err)
		require.Equal(t, username, u.Username())
		return nil
	})
	require.NoError(t, err)
}

func TestRemoveSweepsSessionIndex(t *testing.T) {
	tp, cleanup := fixture.AuthTransactionProvider(t)
	defer cleanup()

	username, err := authdomain.NewUsernameFromString("username")
	require.NoError(t, err)
	password, err := authdomain.NewPasswordHash([]byte("hash"))
	require.NoError(t, err)
	tokenA, err := authdomain.NewAccessTokenFromString("a")
	require.NoError(t, err)
	tokenB, err := authdomain.NewAccessTokenFromString("b")
	require.NoError(t, err)

	now := time.Now()
	user := authdomain.NewUserFromDatabase(
		username, password, false, &now, &now,
		[]authdomain.Session{
			authdomain.NewSession(tokenA, now),
			authdomain.NewSession(tokenB, now),
		},
	)

	err = tp.Write(func(r *app.TransactableRepositories) error {
		return r.Users.Put(user)
	})
	require.NoError(t, err)

	err = tp.Write(func(r *app.TransactableRepositories) error {
		return r.Users.Remove(username)
	})
	require.NoError(t, err)

	err = tp.Read(func(r *app.TransactableRepositories) error {
		_, err := r.Users.GetByToken(tokenA)
		require.ErrorIs(t, err, app.ErrNotFound)
		_, err = r.Users.GetByToken(tokenB)
		require.ErrorIs(t, err, app.ErrNotFound)
		return nil
	})
	require.NoError(t, err)
}

func TestPutIndexIsolatedAcrossUsers(t *testing.T) {
	tp, cleanup := fixture.AuthTransactionProvider(t)
	defer cleanup()

	password, err := authdomain.NewPasswordHash([]byte("hash"))
	require.NoError(t, err)

	usernameA, err := authdomain.NewUsernameFromString("ua")
	require.NoError(t, err)
	usernameB, err := authdomain.NewUsernameFromString("ub")
	require.NoError(t, err)

	tokenA, err := authdomain.NewAccessTokenFromString("a")
	require.NoError(t, err)
	tokenB, err := authdomain.NewAccessTokenFromString("b")
	require.NoError(t, err)

	now := time.Now()

	err = tp.Write(func(r *app.TransactableRepositories) error {
		if err := r.Users.Put(authdomain.NewUserFromDatabase(
			usernameA, password, false, &now, &now,
			[]authdomain.Session{authdomain.NewSession(tokenA, now)},
		)); err != nil {
			return err
		}
		return r.Users.Put(authdomain.NewUserFromDatabase(
			usernameB, password, false, &now, &now,
			[]authdomain.Session{authdomain.NewSession(tokenB, now)},
		))
	})
	require.NoError(t, err)

	err = tp.Write(func(r *app.TransactableRepositories) error {
		return r.Users.Remove(usernameA)
	})
	require.NoError(t, err)

	err = tp.Read(func(r *app.TransactableRepositories) error {
		_, err := r.Users.GetByToken(tokenA)
		require.ErrorIs(t, err, app.ErrNotFound)

		u, err := r.Users.GetByToken(tokenB)
		require.NoError(t, err)
		require.Equal(t, usernameB, u.Username())
		return nil
	})
	require.NoError(t, err)
}

func TestUserWithMissingCreatedAndLastSeen(t *testing.T) {
	tp, cleanup := fixture.AuthTransactionProvider(t)
	defer cleanup()

	username, err := authdomain.NewUsernameFromString("username")
	require.NoError(t, err)
	password, err := authdomain.NewPasswordHash([]byte("hash"))
	require.NoError(t, err)

	user := authdomain.NewUserFromDatabase(username, password, false, nil, nil, nil)

	err = tp.Write(func(r *app.TransactableRepositories) error {
		return r.Users.Put(user)
	})
	require.NoError(t, err)

	err = tp.Read(func(r *app.TransactableRepositories) error {
		u, err := r.Users.Get(username)
		if err != nil {
			return errors.Wrap(err, "get failed")
		}

		require.Nil(t, u.Created())
		require.Nil(t, u.LastSeen())

		return nil
	})
	require.NoError(t, err)
}
