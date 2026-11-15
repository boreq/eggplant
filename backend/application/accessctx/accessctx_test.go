package accessctx_test

import (
	"testing"
	"time"

	"github.com/boreq/eggplant/application/accessctx"
	authdomain "github.com/boreq/eggplant/domain/auth"
	remotedomain "github.com/boreq/eggplant/domain/remote"
	"github.com/stretchr/testify/require"
)

func TestUserAccessContext_StringDoesNotContainToken(t *testing.T) {
	username, err := authdomain.NewUsernameFromString("test")
	require.NoError(t, err)

	passwordHash, err := authdomain.NewPasswordHash([]byte("hash"))
	require.NoError(t, err)

	token, err := authdomain.NewAccessToken()
	require.NoError(t, err)

	user := authdomain.NewUser(username, passwordHash, true, time.Now())
	accessCtx := accessctx.NewUserAccessContext(user, token)

	require.Equal(t, "user(username=test, administrator=true)", accessCtx.String())
	require.NotContains(t, accessCtx.String(), token.String())
}

func TestRemoteInstanceAccessContext_String(t *testing.T) {
	id, err := remotedomain.NewRemoteInstanceIDFromString("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	require.NoError(t, err)

	accessCtx := accessctx.NewRemoteInstanceAccessContext(id)

	require.Equal(t, "remote_instance(id=01ARZ3NDEKTSV4RRFFQ69G5FAV)", accessCtx.String())
}
