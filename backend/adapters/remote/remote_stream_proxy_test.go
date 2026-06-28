package remote_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	remoteadapter "github.com/boreq/eggplant/adapters/remote"
	appremote "github.com/boreq/eggplant/application/remote"
	"github.com/boreq/eggplant/domain/crockford"
	remotedomain "github.com/boreq/eggplant/domain/remote"
	"github.com/boreq/eggplant/internal/fixture"
	"github.com/boreq/errors"
	"github.com/stretchr/testify/require"
)

func TestStreamProxyForwardsToPeerWithAuthToken(t *testing.T) {
	tp, _, cleanup := fixture.RemoteTransactionProvider(t)
	defer cleanup()

	var gotPath, gotAuth string
	peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("playlist-body"))
	}))
	defer peer.Close()

	instance, authToken := saveInstanceWithAuthToken(t, tp, peer.URL)

	proxy := remoteadapter.NewStreamProxy(tp)

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/remote/"+instance.Id().String()+"/track/TRACK/stream/STREAM/playlist", nil)
	require.NoError(t, proxy.Proxy(rec, req, instance.Id(), "/TRACK/stream/STREAM/playlist"))

	require.Equal(t, "/api/track/TRACK/stream/STREAM/playlist", gotPath)
	require.Equal(t, "Bearer "+crockford.Encode(authToken.Bytes()), gotAuth)

	res := rec.Result()
	defer res.Body.Close()
	require.Equal(t, http.StatusOK, res.StatusCode)
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	require.Equal(t, "playlist-body", string(body))
}

func TestStreamProxyReturnsNotFoundForUnknownInstance(t *testing.T) {
	tp, _, cleanup := fixture.RemoteTransactionProvider(t)
	defer cleanup()

	proxy := remoteadapter.NewStreamProxy(tp)

	id, err := remotedomain.NewRemoteInstanceID()
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/remote/"+id.String()+"/track/T/stream/S/playlist", nil)
	err = proxy.Proxy(rec, req, id, "/T/stream/S/playlist")
	require.ErrorIs(t, err, appremote.ErrNotFound)
}

func saveInstanceWithAuthToken(t *testing.T, tp appremote.TransactionProvider, address string) (*remotedomain.RemoteInstance, remotedomain.AuthToken) {
	t.Helper()

	id, err := remotedomain.NewRemoteInstanceID()
	require.NoError(t, err)

	addr, err := remotedomain.NewRemoteInstanceAddress(address)
	require.NoError(t, err)

	localPairingToken, err := remotedomain.NewPairingToken()
	require.NoError(t, err)

	instance := remotedomain.NewRemoteInstance(id, addr, localPairingToken.Hash())

	remotePairingToken, err := remotedomain.NewPairingToken()
	require.NoError(t, err)
	require.NoError(t, instance.SetRemotePairingToken(remotePairingToken))

	_, err = instance.IssueLocalAuthToken()
	require.NoError(t, err)

	authToken, err := remotedomain.NewAuthToken()
	require.NoError(t, err)
	require.NoError(t, instance.SetRemoteAuthToken(localPairingToken, authToken))

	require.NoError(t, tp.Write(func(r *appremote.TransactableRepositories) error {
		return errors.Wrap(r.RemoteInstances.Save(instance), "save failed")
	}))

	return instance, authToken
}
