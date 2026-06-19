package componenttest

import (
	"context"
	"net/http"
	"testing"
	"time"

	remoteadapter "github.com/boreq/eggplant/adapters/remote"
	remotedomain "github.com/boreq/eggplant/domain/remote"
	"github.com/boreq/eggplant/entrypoints/http/openapi"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
)

func TestServiceRemotePairing(t *testing.T) {
	ctx := context.Background()

	instanceA := newTestService(t)
	instanceB := newTestService(t)

	adminA := authedAs(registerAdminAndLogin(t, instanceA))
	adminB := authedAs(registerAdminAndLogin(t, instanceB))

	startedA, err := instanceA.client.AddRemoteWithResponse(ctx, openapi.AddRemoteJSONRequestBody{
		Url: instanceB.baseURL,
	}, adminA)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, startedA.StatusCode())
	require.NotNil(t, startedA.JSON200)
	require.NotEmpty(t, startedA.JSON200.Id)
	require.NotEmpty(t, startedA.JSON200.LocalPairingToken)

	startedB, err := instanceB.client.AddRemoteWithResponse(ctx, openapi.AddRemoteJSONRequestBody{
		Url: instanceA.baseURL,
	}, adminB)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, startedB.StatusCode())
	require.NotNil(t, startedB.JSON200)

	setA, err := instanceA.client.SetRemotePairingTokenWithResponse(ctx, startedA.JSON200.Id, openapi.SetRemotePairingTokenJSONRequestBody{
		PeerToken: startedB.JSON200.LocalPairingToken,
	}, adminA)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, setA.StatusCode())

	setB, err := instanceB.client.SetRemotePairingTokenWithResponse(ctx, startedB.JSON200.Id, openapi.SetRemotePairingTokenJSONRequestBody{
		PeerToken: startedA.JSON200.LocalPairingToken,
	}, adminB)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, setB.StatusCode())

	// Setting the peer token publishes an event; each instance's pub/sub then
	// delivers its auth token to the peer, which stores it.
	idA := mustRemoteInstanceID(t, startedA.JSON200.Id)
	idB := mustRemoteInstanceID(t, startedB.JSON200.Id)
	require.Eventually(t, func() bool {
		return remoteAuthTokenSet(instanceA, idA) && remoteAuthTokenSet(instanceB, idB)
	}, 10*time.Second, 50*time.Millisecond)

	// Completing the pairing publishes a RemotePaired event whose handler
	// healthchecks the instance right away, so it becomes HEALTHY without
	// waiting for the periodic healthcheck timer.
	require.Eventually(t, func() bool {
		listed, err := instanceA.client.ListRemotesWithResponse(ctx, adminA)
		if err != nil || listed.JSON200 == nil || len(*listed.JSON200) != 1 {
			return false
		}
		return (*listed.JSON200)[0].Status == openapi.RemoteInstanceStatusHEALTHY
	}, 10*time.Second, 50*time.Millisecond)

	listed, err := instanceA.client.ListRemotesWithResponse(ctx, adminA)
	require.NoError(t, err)
	require.NotNil(t, listed.JSON200)
	require.Len(t, *listed.JSON200, 1)
	require.Equal(t, instanceB.baseURL, (*listed.JSON200)[0].Address)
	require.Equal(t, openapi.RemoteInstanceStatusHEALTHY, (*listed.JSON200)[0].Status)
	require.NotNil(t, (*listed.JSON200)[0].LastHealthcheckStatus)
	require.Equal(t, openapi.RemoteInstanceLastHealthcheckStatusALIVE, *(*listed.JSON200)[0].LastHealthcheckStatus)
	require.NotNil(t, (*listed.JSON200)[0].LastHealthcheckAt)
}

func remoteAuthTokenSet(ts *testService, id remotedomain.RemoteInstanceID) bool {
	var set bool
	_ = ts.db.View(func(tx *bolt.Tx) error {
		r, err := remoteadapter.NewRemoteInstanceRepository(tx)
		if err != nil {
			return err
		}
		instance, err := r.GetByID(id)
		if err != nil {
			return err
		}
		_, set = instance.RemoteAuthToken()
		return nil
	})
	return set
}

func mustRemoteInstanceID(t *testing.T, s string) remotedomain.RemoteInstanceID {
	id, err := remotedomain.NewRemoteInstanceIDFromString(s)
	require.NoError(t, err)
	return id
}
