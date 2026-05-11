package auth_test

import (
	"testing"
	"time"

	"github.com/boreq/eggplant/adapters/auth"
	app "github.com/boreq/eggplant/application/auth"
	"github.com/boreq/eggplant/internal/fixture"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
)

func TestRemoveOldInvitationsWhenSavingNewOnes(t *testing.T) {
	db, cleanup := fixture.Bolt(t)
	defer cleanup()

	invitation1 := app.Invitation{
		Token:   "1",
		Created: time.Now().Add(-72 * time.Hour),
	}

	invitation2 := app.Invitation{
		Token:   "2",
		Created: time.Now(),
	}

	invitation3 := app.Invitation{
		Token:   "3",
		Created: time.Now(),
	}

	err := db.Update(func(tx *bolt.Tx) error {
		r, err := auth.NewInvitationRepository(tx)
		require.NoError(t, err)

		for _, invitation := range []app.Invitation{invitation1, invitation2, invitation3} {
			err := r.Put(invitation)
			require.NoError(t, err)
		}

		return nil
	})
	require.NoError(t, err)

	err = db.View(func(tx *bolt.Tx) error {
		r, err := auth.NewInvitationRepository(tx)
		require.NoError(t, err)

		_, err = r.Get(invitation1.Token)
		require.ErrorIs(t, err, app.ErrNotFound)

		_, err = r.Get(invitation2.Token)
		require.NoError(t, err)

		_, err = r.Get(invitation3.Token)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)
}
