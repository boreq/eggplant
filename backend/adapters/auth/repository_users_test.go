package auth_test

import (
	"testing"
	"time"

	"github.com/boreq/eggplant/adapters/auth"
	authdomain "github.com/boreq/eggplant/domain/auth"
	"github.com/boreq/eggplant/internal/fixture"
	"github.com/boreq/errors"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
)

func TestRemoveOldSessionsWhenSavingUsers(t *testing.T) {
	db, cleanup := fixture.Bolt(t)
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

	user := authdomain.NewUser(
		username,
		password,
		false,
		time.Now(),
		time.Now(),
		[]authdomain.Session{session1, session2},
	)

	err = db.Update(func(tx *bolt.Tx) error {
		r, err := auth.NewUserRepository(tx)
		if err != nil {
			return err
		}
		return r.Put(user)
	})
	require.NoError(t, err)

	err = db.View(func(tx *bolt.Tx) error {
		r, err := auth.NewUserRepository(tx)
		if err != nil {
			return err
		}

		u, err := r.Get(username)
		if err != nil {
			return errors.Wrap(err, "get failed")
		}

		require.Len(t, u.Sessions(), 1)
		require.Equal(t, tokenRecent, u.Sessions()[0].Token())

		return nil
	})
	require.NoError(t, err)
}
