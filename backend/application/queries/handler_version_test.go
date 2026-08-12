package queries_test

import (
	"testing"
	"time"

	"github.com/boreq/eggplant/application/accessctx"
	"github.com/boreq/eggplant/application/queries"
	authdomain "github.com/boreq/eggplant/domain/auth"
	remotedomain "github.com/boreq/eggplant/domain/remote"
	"github.com/stretchr/testify/require"
)

func TestVersion(t *testing.T) {
	testCases := []struct {
		name      string
		accessCtx accessctx.AccessContext
		expectErr error
	}{
		{
			name:      "administrator",
			accessCtx: userAccessContext(t, true),
			expectErr: nil,
		},
		{
			name:      "regular_user",
			accessCtx: userAccessContext(t, false),
			expectErr: accessctx.ErrPermissionDenied,
		},
		{
			name:      "anonymous",
			accessCtx: accessctx.NewAnonymousAccessContext(),
			expectErr: accessctx.ErrPermissionDenied,
		},
		{
			name:      "remote_instance",
			accessCtx: accessctx.NewRemoteInstanceAccessContext(remoteInstanceID(t)),
			expectErr: nil,
		},
		{
			name:      "command_line",
			accessCtx: accessctx.NewCommandLineAccessContext(),
			expectErr: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			version, err := queries.NewVersionHandler("backend-v1.2.3", "frontend-v1.2.3").Execute(testCase.accessCtx)
			if testCase.expectErr != nil {
				require.ErrorIs(t, err, testCase.expectErr)
				require.Equal(t, "", version.Backend())
				require.Equal(t, "", version.Frontend())
				return
			}
			require.NoError(t, err)
			require.Equal(t, "backend-v1.2.3", version.Backend())
			require.Equal(t, "frontend-v1.2.3", version.Frontend())
		})
	}
}

func userAccessContext(t *testing.T, administrator bool) accessctx.AccessContext {
	t.Helper()

	username, err := authdomain.NewUsernameFromString("username")
	require.NoError(t, err)

	token, err := authdomain.NewAccessToken()
	require.NoError(t, err)

	user := authdomain.NewUser(username, authdomain.PasswordHash{}, administrator, time.Now())
	return accessctx.NewUserAccessContext(user, token)
}

func remoteInstanceID(t *testing.T) remotedomain.RemoteInstanceID {
	t.Helper()
	id, err := remotedomain.NewRemoteInstanceID()
	require.NoError(t, err)
	return id
}
