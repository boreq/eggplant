package auth_test

import (
	"testing"
	"time"

	"github.com/boreq/eggplant/adapters/auth"
	authdomain "github.com/boreq/eggplant/domain/auth"
	"github.com/stretchr/testify/require"
)

func TestLastSeenUpdater_PopUpdatesReturnsBufferedUpdates(t *testing.T) {
	u := auth.NewLastSeenUpdater()

	username := mustUsername(t, "username")
	tokenA := mustToken(t, "a")
	tokenB := mustToken(t, "b")

	t1 := time.Now()
	t2 := t1.Add(time.Second)

	u.Update(username, tokenA, t1)
	u.Update(username, tokenB, t2)

	updates := u.PopUpdates()
	require.Len(t, updates, 1)

	update := updates[0]
	require.Equal(t, username, update.Username)
	require.True(t, update.LastSeen.Equal(t2))
	require.True(t, update.Sessions[tokenA].Equal(t1))
	require.True(t, update.Sessions[tokenB].Equal(t2))
}

func TestLastSeenUpdater_KeepsLatestTimestamp(t *testing.T) {
	u := auth.NewLastSeenUpdater()

	username := mustUsername(t, "username")
	token := mustToken(t, "a")

	newer := time.Now()
	older := newer.Add(-time.Second)

	u.Update(username, token, newer)
	u.Update(username, token, older)

	updates := u.PopUpdates()
	require.Len(t, updates, 1)
	require.True(t, updates[0].LastSeen.Equal(newer))
	require.True(t, updates[0].Sessions[token].Equal(newer))
}

func TestLastSeenUpdater_PopUpdatesClearsBuffer(t *testing.T) {
	u := auth.NewLastSeenUpdater()

	u.Update(mustUsername(t, "username"), mustToken(t, "a"), time.Now())

	require.Len(t, u.PopUpdates(), 1)
	require.Nil(t, u.PopUpdates())
}

func mustUsername(t *testing.T, s string) authdomain.Username {
	username, err := authdomain.NewUsernameFromString(s)
	require.NoError(t, err)
	return username
}

func mustToken(t *testing.T, s string) authdomain.AccessToken {
	token, err := authdomain.NewAccessTokenFromString(s)
	require.NoError(t, err)
	return token
}
