package auth_test

import (
	"testing"
	"time"

	"github.com/boreq/eggplant/adapters/auth"
	app "github.com/boreq/eggplant/application/auth"
	"github.com/boreq/eggplant/internal/fixture"
	"github.com/boreq/errors"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
)

func TestRemoveOldSessionsWhenSavingUsers(t *testing.T) {
	db, cleanup := fixture.Bolt(t)
	defer cleanup()

	session1 := app.Session{
		Token:    "a",
		LastSeen: time.Now().Add(-2 * 365 * 24 * time.Hour),
	}

	session2 := app.Session{
		Token:    "a",
		LastSeen: time.Now().Add(-10 * 24 * time.Hour),
	}

	err := db.Update(func(tx *bolt.Tx) error {
		r, err := auth.NewUserRepository(tx)
		if err != nil {
			return err
		}

		return r.Put(app.User{
			Username: "username",
			Sessions: []app.Session{
				session1,
				session2,
			},
		})
	})
	require.NoError(t, err)

	err = db.View(func(tx *bolt.Tx) error {
		r, err := auth.NewUserRepository(tx)
		if err != nil {
			return err
		}

		u, err := r.Get("username")
		if err != nil {
			return errors.Wrap(err, "get failed")
		}

		require.Len(t, u.Sessions, 1)
		require.Equal(t, session2.Token, u.Sessions[0].Token)

		return nil
	})
	require.NoError(t, err)
}
