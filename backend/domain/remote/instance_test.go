package remote_test

import (
	"testing"
	"time"

	remotedomain "github.com/boreq/eggplant/domain/remote"
	"github.com/stretchr/testify/require"
)

func newTestID(t *testing.T) remotedomain.RemoteInstanceID {
	id, err := remotedomain.NewRemoteInstanceID()
	require.NoError(t, err)
	return id
}

func newTestAddress(t *testing.T) remotedomain.RemoteInstanceAddress {
	address, err := remotedomain.NewRemoteInstanceAddress("https://example.com")
	require.NoError(t, err)
	return address
}

func newTestPairingTokenHash(t *testing.T) remotedomain.HashedPairingToken {
	token, err := remotedomain.NewPairingToken()
	require.NoError(t, err)
	return token.Hash()
}

func newTestInstance(t *testing.T) (*remotedomain.RemoteInstance, remotedomain.PairingToken) {
	localPairingToken, err := remotedomain.NewPairingToken()
	require.NoError(t, err)
	instance := remotedomain.NewRemoteInstance(newTestID(t), newTestAddress(t), localPairingToken.Hash())
	return instance, localPairingToken
}

func TestNewRemoteInstance(t *testing.T) {
	id := newTestID(t)
	address := newTestAddress(t)

	token, err := remotedomain.NewPairingToken()
	require.NoError(t, err)

	instance := remotedomain.NewRemoteInstance(id, address, token.Hash())

	require.Equal(t, id, instance.Id())
	require.Equal(t, address, instance.Address())

	hash, ok := instance.LocalPairingTokenHash()
	require.True(t, ok, "a fresh instance has a local pairing token hash")
	require.Equal(t, token.Hash().Bytes(), hash.Bytes())

	_, ok = instance.RemotePairingToken()
	require.False(t, ok)

	_, ok = instance.LocalAuthTokenHash()
	require.False(t, ok)

	_, ok = instance.RemoteAuthToken()
	require.False(t, ok)

	require.Empty(t, instance.PopEvents())
}

func TestSetRemotePairingToken(t *testing.T) {
	instance := remotedomain.NewRemoteInstance(newTestID(t), newTestAddress(t), newTestPairingTokenHash(t))

	token, err := remotedomain.NewPairingToken()
	require.NoError(t, err)

	err = instance.SetRemotePairingToken(token)
	require.NoError(t, err)

	got, ok := instance.RemotePairingToken()
	require.True(t, ok)
	require.Equal(t, token, got)

	events := instance.PopEvents()
	require.Len(t, events, 1)
	require.IsType(t, remotedomain.RemotePairingTokenSet{}, events[0])
	require.Equal(t, instance.Id(), events[0].(remotedomain.RemotePairingTokenSet).RemoteInstanceID)
}

func TestSetRemotePairingTokenIsIdempotent(t *testing.T) {
	instance := remotedomain.NewRemoteInstance(newTestID(t), newTestAddress(t), newTestPairingTokenHash(t))

	token, err := remotedomain.NewPairingToken()
	require.NoError(t, err)

	require.NoError(t, instance.SetRemotePairingToken(token))
	require.Len(t, instance.PopEvents(), 1)

	require.NoError(t, instance.SetRemotePairingToken(token))
	require.Empty(t, instance.PopEvents())

	got, ok := instance.RemotePairingToken()
	require.True(t, ok)
	require.Equal(t, token, got)
}

func TestSetRemotePairingTokenSameTokenSucceedsAfterPairing(t *testing.T) {
	instance, localPairingToken := newTestInstance(t)

	pairingToken, err := remotedomain.NewPairingToken()
	require.NoError(t, err)

	require.NoError(t, instance.SetRemotePairingToken(pairingToken))
	instance.PopEvents()

	authToken, err := remotedomain.NewAuthToken()
	require.NoError(t, err)

	require.NoError(t, instance.SetRemoteAuthToken(localPairingToken, authToken))

	require.NoError(t, instance.SetRemotePairingToken(pairingToken))
	require.Empty(t, instance.PopEvents())
}

func TestSetRemotePairingTokenSucceedsAfterRemoteAuthTokenReceived(t *testing.T) {
	instance, localPairingToken := newTestInstance(t)

	authToken, err := remotedomain.NewAuthToken()
	require.NoError(t, err)

	err = instance.SetRemoteAuthToken(localPairingToken, authToken)
	require.NoError(t, err)

	pairingToken, err := remotedomain.NewPairingToken()
	require.NoError(t, err)

	err = instance.SetRemotePairingToken(pairingToken)
	require.NoError(t, err)
}

func TestSetRemotePairingTokenFailsAfterLocalAuthTokenIssued(t *testing.T) {
	instance, _ := newTestInstance(t)

	first, err := remotedomain.NewPairingToken()
	require.NoError(t, err)
	require.NoError(t, instance.SetRemotePairingToken(first))

	_, err = instance.IssueLocalAuthToken()
	require.NoError(t, err)

	second, err := remotedomain.NewPairingToken()
	require.NoError(t, err)

	err = instance.SetRemotePairingToken(second)
	require.EqualError(t, err, "the local auth token has already been issued for this remote instance")
}

func TestPairingCompletesWhenLocalAuthTokenIssuedLast(t *testing.T) {
	instance, localPairingToken := newTestInstance(t)

	authToken, err := remotedomain.NewAuthToken()
	require.NoError(t, err)
	require.NoError(t, instance.SetRemoteAuthToken(localPairingToken, authToken))
	require.Empty(t, instance.PopEvents())

	_, err = instance.IssueLocalAuthToken()
	require.NoError(t, err)

	requireRemotePairedRecorded(t, instance)
}

func TestPairingCompletesWhenRemoteAuthTokenSetLast(t *testing.T) {
	instance, localPairingToken := newTestInstance(t)

	_, err := instance.IssueLocalAuthToken()
	require.NoError(t, err)
	require.Empty(t, instance.PopEvents())

	authToken, err := remotedomain.NewAuthToken()
	require.NoError(t, err)
	require.NoError(t, instance.SetRemoteAuthToken(localPairingToken, authToken))

	requireRemotePairedRecorded(t, instance)
}

func requireRemotePairedRecorded(t *testing.T, instance *remotedomain.RemoteInstance) {
	events := instance.PopEvents()
	require.Len(t, events, 1)
	paired, ok := events[0].(remotedomain.RemotePaired)
	require.True(t, ok)
	require.Equal(t, instance.Id(), paired.RemoteInstanceID)
}

func TestSetRemoteAuthToken(t *testing.T) {
	instance, localPairingToken := newTestInstance(t)

	authToken, err := remotedomain.NewAuthToken()
	require.NoError(t, err)

	err = instance.SetRemoteAuthToken(localPairingToken, authToken)
	require.NoError(t, err)

	got, ok := instance.RemoteAuthToken()
	require.True(t, ok)
	require.Equal(t, authToken, got)

	_, ok = instance.LocalPairingTokenHash()
	require.False(t, ok)
}

func TestSetRemoteAuthTokenIsIdempotent(t *testing.T) {
	instance, localPairingToken := newTestInstance(t)

	authToken, err := remotedomain.NewAuthToken()
	require.NoError(t, err)

	require.NoError(t, instance.SetRemoteAuthToken(localPairingToken, authToken))

	require.NoError(t, instance.SetRemoteAuthToken(localPairingToken, authToken))

	got, ok := instance.RemoteAuthToken()
	require.True(t, ok)
	require.Equal(t, authToken, got)

	_, ok = instance.LocalPairingTokenHash()
	require.False(t, ok)
}

func TestSetRemoteAuthTokenRejectsWrongLocalPairingToken(t *testing.T) {
	instance, _ := newTestInstance(t)

	wrongPairingToken, err := remotedomain.NewPairingToken()
	require.NoError(t, err)

	authToken, err := remotedomain.NewAuthToken()
	require.NoError(t, err)

	err = instance.SetRemoteAuthToken(wrongPairingToken, authToken)
	require.EqualError(t, err, "the local pairing token does not match")

	_, ok := instance.RemoteAuthToken()
	require.False(t, ok)

	_, ok = instance.LocalPairingTokenHash()
	require.True(t, ok)
}

func TestSetRemoteAuthTokenFailsWhenLocalPairingTokenHashAlreadyCleared(t *testing.T) {
	instance, localPairingToken := newTestInstance(t)

	authToken, err := remotedomain.NewAuthToken()
	require.NoError(t, err)

	require.NoError(t, instance.SetRemoteAuthToken(localPairingToken, authToken))

	otherAuthToken, err := remotedomain.NewAuthToken()
	require.NoError(t, err)

	err = instance.SetRemoteAuthToken(localPairingToken, otherAuthToken)
	require.EqualError(t, err, "the local pairing token hash is not set")

	got, ok := instance.RemoteAuthToken()
	require.True(t, ok)
	require.Equal(t, authToken, got)
}

func TestStatusPairingUntilBothDirectionsEstablished(t *testing.T) {
	instance, localPairingToken := newTestInstance(t)
	require.Equal(t, remotedomain.RemoteInstanceStatusPairing, instance.Status())

	_, err := instance.IssueLocalAuthToken()
	require.NoError(t, err)
	require.Equal(t, remotedomain.RemoteInstanceStatusPairing, instance.Status())

	authToken, err := remotedomain.NewAuthToken()
	require.NoError(t, err)
	require.NoError(t, instance.SetRemoteAuthToken(localPairingToken, authToken))

	require.Equal(t, remotedomain.RemoteInstanceStatusDead, instance.Status())
}

func TestStatusPairingWithRemoteAuthTokenOnly(t *testing.T) {
	instance, localPairingToken := newTestInstance(t)

	authToken, err := remotedomain.NewAuthToken()
	require.NoError(t, err)
	require.NoError(t, instance.SetRemoteAuthToken(localPairingToken, authToken))

	require.Equal(t, remotedomain.RemoteInstanceStatusPairing, instance.Status())
}

func TestStatusHealthyAndDeadFollowHealthcheck(t *testing.T) {
	instance, localPairingToken := newTestInstance(t)

	_, err := instance.IssueLocalAuthToken()
	require.NoError(t, err)
	authToken, err := remotedomain.NewAuthToken()
	require.NoError(t, err)
	require.NoError(t, instance.SetRemoteAuthToken(localPairingToken, authToken))

	require.NoError(t, instance.RecordHealthcheck(remotedomain.HealthcheckStatusAlive, time.Now()))
	require.Equal(t, remotedomain.RemoteInstanceStatusHealthy, instance.Status())

	require.NoError(t, instance.RecordHealthcheck(remotedomain.HealthcheckStatusDead, time.Now()))
	require.Equal(t, remotedomain.RemoteInstanceStatusDead, instance.Status())
}

func TestIssueLocalAuthToken(t *testing.T) {
	instance := remotedomain.NewRemoteInstance(newTestID(t), newTestAddress(t), newTestPairingTokenHash(t))

	_, ok := instance.LocalAuthTokenHash()
	require.False(t, ok)

	token, err := instance.IssueLocalAuthToken()
	require.NoError(t, err)

	hash, ok := instance.LocalAuthTokenHash()
	require.True(t, ok)
	require.Equal(t, token.Hash().Bytes(), hash.Bytes())
}

func TestNewRemoteInstanceFromDatabase(t *testing.T) {
	id := newTestID(t)
	address := newTestAddress(t)

	localPairingToken, err := remotedomain.NewPairingToken()
	require.NoError(t, err)
	localPairingTokenHash := localPairingToken.Hash()

	remotePairingToken, err := remotedomain.NewPairingToken()
	require.NoError(t, err)

	authToken, err := remotedomain.NewAuthToken()
	require.NoError(t, err)
	localAuthTokenHash := authToken.Hash()

	remoteAuthToken, err := remotedomain.NewAuthToken()
	require.NoError(t, err)

	healthcheck, err := remotedomain.NewHealthcheck(remotedomain.HealthcheckStatusAlive, time.Now())
	require.NoError(t, err)

	instance := remotedomain.NewRemoteInstanceFromDatabase(
		id,
		address,
		&localPairingTokenHash,
		&localAuthTokenHash,
		&remotePairingToken,
		&remoteAuthToken,
		&healthcheck,
	)

	require.Equal(t, id, instance.Id())
	require.Equal(t, address, instance.Address())

	gotLocalPairingHash, ok := instance.LocalPairingTokenHash()
	require.True(t, ok)
	require.Equal(t, localPairingTokenHash, gotLocalPairingHash)

	gotRemotePairing, ok := instance.RemotePairingToken()
	require.True(t, ok)
	require.Equal(t, remotePairingToken, gotRemotePairing)

	gotHash, ok := instance.LocalAuthTokenHash()
	require.True(t, ok)
	require.Equal(t, localAuthTokenHash, gotHash)

	gotRemoteAuth, ok := instance.RemoteAuthToken()
	require.True(t, ok)
	require.Equal(t, remoteAuthToken, gotRemoteAuth)

	gotHealthcheck, ok := instance.LastHealthcheck()
	require.True(t, ok)
	require.Equal(t, healthcheck, gotHealthcheck)
}
