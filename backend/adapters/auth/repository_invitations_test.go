package auth_test

import (
	"testing"
	"time"

	"github.com/boreq/eggplant/adapters/auth"
	app "github.com/boreq/eggplant/application/auth"
	authdomain "github.com/boreq/eggplant/domain/auth"
	"github.com/boreq/eggplant/internal/fixture"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
)

func TestRemoveOldInvitationsWhenSavingNewOnes(t *testing.T) {
	db, cleanup := fixture.Bolt(t)
	defer cleanup()

	token1, err := authdomain.NewInvitationTokenFromString("1")
	require.NoError(t, err)
	token2, err := authdomain.NewInvitationTokenFromString("2")
	require.NoError(t, err)
	token3, err := authdomain.NewInvitationTokenFromString("3")
	require.NoError(t, err)

	invitation1 := authdomain.NewInvitation(token1, time.Now().Add(-72*time.Hour))
	invitation2 := authdomain.NewInvitation(token2, time.Now())
	invitation3 := authdomain.NewInvitation(token3, time.Now())

	err = db.Update(func(tx *bolt.Tx) error {
		r, err := auth.NewInvitationRepository(tx)
		require.NoError(t, err)

		for _, invitation := range []authdomain.Invitation{invitation1, invitation2, invitation3} {
			err := r.Put(invitation)
			require.NoError(t, err)
		}

		return nil
	})
	require.NoError(t, err)

	err = db.View(func(tx *bolt.Tx) error {
		r, err := auth.NewInvitationRepository(tx)
		require.NoError(t, err)

		_, err = r.Get(invitation1.Token())
		require.ErrorIs(t, err, app.ErrNotFound)

		_, err = r.Get(invitation2.Token())
		require.NoError(t, err)

		_, err = r.Get(invitation3.Token())
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)
}
