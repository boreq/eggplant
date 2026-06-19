package remote_test

import (
	"testing"
	"time"

	appremote "github.com/boreq/eggplant/application/remote"
	remotedomain "github.com/boreq/eggplant/domain/remote"
	"github.com/boreq/eggplant/internal/fixture"
	"github.com/stretchr/testify/require"
)

func TestSaveAndGetByIDRoundTripsAllFields(t *testing.T) {
	tp, _, cleanup := fixture.RemoteTransactionProvider(t)
	defer cleanup()

	instance, localPairingToken := newTestInstanceWithPairingToken(t)

	remotePairingToken, err := remotedomain.NewPairingToken()
	require.NoError(t, err)
	require.NoError(t, instance.SetRemotePairingToken(remotePairingToken))

	localAuthToken, err := instance.IssueLocalAuthToken()
	require.NoError(t, err)

	remoteAuthToken, err := remotedomain.NewAuthToken()
	require.NoError(t, err)
	require.NoError(t, instance.SetRemoteAuthToken(localPairingToken, remoteAuthToken))

	healthcheckTime := time.Date(2026, 6, 20, 10, 11, 12, 123456789, time.UTC)
	require.NoError(t, instance.RecordHealthcheck(remotedomain.HealthcheckStatusAlive, healthcheckTime))

	require.NoError(t, tp.Write(func(r *appremote.TransactableRepositories) error {
		return r.RemoteInstances.Save(instance)
	}))

	var got *remotedomain.RemoteInstance
	require.NoError(t, tp.Read(func(r *appremote.TransactableRepositories) error {
		var err error
		got, err = r.RemoteInstances.GetByID(instance.Id())
		return err
	}))

	require.Equal(t, instance.Id().String(), got.Id().String())
	require.Equal(t, instance.Address().String(), got.Address().String())

	// SetRemoteAuthToken clears the local pairing token hash.
	_, ok := got.LocalPairingTokenHash()
	require.False(t, ok)

	gotLocalAuthHash, ok := got.LocalAuthTokenHash()
	require.True(t, ok)
	require.Equal(t, localAuthToken.Hash().Bytes(), gotLocalAuthHash.Bytes())

	gotRemotePairing, ok := got.RemotePairingToken()
	require.True(t, ok)
	require.Equal(t, remotePairingToken.Bytes(), gotRemotePairing.Bytes())

	gotRemoteAuth, ok := got.RemoteAuthToken()
	require.True(t, ok)
	require.Equal(t, remoteAuthToken.Bytes(), gotRemoteAuth.Bytes())

	gotHealthcheck, ok := got.LastHealthcheck()
	require.True(t, ok)
	require.Equal(t, remotedomain.HealthcheckStatusAlive, gotHealthcheck.Status())
	require.True(t, healthcheckTime.Equal(gotHealthcheck.At()))
}

func TestSaveAndGetByIDRoundTripsMinimalInstance(t *testing.T) {
	tp, _, cleanup := fixture.RemoteTransactionProvider(t)
	defer cleanup()

	instance := newTestInstance(t)

	require.NoError(t, tp.Write(func(r *appremote.TransactableRepositories) error {
		return r.RemoteInstances.Save(instance)
	}))

	var got *remotedomain.RemoteInstance
	require.NoError(t, tp.Read(func(r *appremote.TransactableRepositories) error {
		var err error
		got, err = r.RemoteInstances.GetByID(instance.Id())
		return err
	}))

	require.Equal(t, instance.Id().String(), got.Id().String())
	require.Equal(t, instance.Address().String(), got.Address().String())

	// A fresh instance has only the local pairing token hash.
	gotLocalPairingHash, ok := got.LocalPairingTokenHash()
	require.True(t, ok)
	originalHash, ok := instance.LocalPairingTokenHash()
	require.True(t, ok)
	require.Equal(t, originalHash.Bytes(), gotLocalPairingHash.Bytes())

	_, ok = got.LocalAuthTokenHash()
	require.False(t, ok)
	_, ok = got.RemotePairingToken()
	require.False(t, ok)
	_, ok = got.RemoteAuthToken()
	require.False(t, ok)
	_, ok = got.LastHealthcheck()
	require.False(t, ok)
}

func TestGetByIDReturnsNotFoundForUnknownID(t *testing.T) {
	tp, _, cleanup := fixture.RemoteTransactionProvider(t)
	defer cleanup()

	id, err := remotedomain.NewRemoteInstanceID()
	require.NoError(t, err)

	require.NoError(t, tp.Read(func(r *appremote.TransactableRepositories) error {
		_, err := r.RemoteInstances.GetByID(id)
		require.ErrorIs(t, err, appremote.ErrNotFound)
		return nil
	}))
}

func TestSaveOverwritesExistingInstanceWithSameID(t *testing.T) {
	tp, _, cleanup := fixture.RemoteTransactionProvider(t)
	defer cleanup()

	instance := newTestInstance(t)

	require.NoError(t, tp.Write(func(r *appremote.TransactableRepositories) error {
		return r.RemoteInstances.Save(instance)
	}))

	require.NoError(t, instance.RecordHealthcheck(remotedomain.HealthcheckStatusDead, time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC)))

	require.NoError(t, tp.Write(func(r *appremote.TransactableRepositories) error {
		return r.RemoteInstances.Save(instance)
	}))

	var all []*remotedomain.RemoteInstance
	require.NoError(t, tp.Read(func(r *appremote.TransactableRepositories) error {
		var err error
		all, err = r.RemoteInstances.GetAll()
		return err
	}))

	require.Len(t, all, 1)
	healthcheck, ok := all[0].LastHealthcheck()
	require.True(t, ok)
	require.Equal(t, remotedomain.HealthcheckStatusDead, healthcheck.Status())
}

func TestGetAllReturnsNilWhenEmpty(t *testing.T) {
	tp, _, cleanup := fixture.RemoteTransactionProvider(t)
	defer cleanup()

	require.NoError(t, tp.Read(func(r *appremote.TransactableRepositories) error {
		all, err := r.RemoteInstances.GetAll()
		require.NoError(t, err)
		require.Empty(t, all)
		return nil
	}))
}

func TestGetAllReturnsEverySavedInstance(t *testing.T) {
	tp, _, cleanup := fixture.RemoteTransactionProvider(t)
	defer cleanup()

	first := newTestInstance(t)
	second := newTestInstance(t)

	require.NoError(t, tp.Write(func(r *appremote.TransactableRepositories) error {
		if err := r.RemoteInstances.Save(first); err != nil {
			return err
		}
		return r.RemoteInstances.Save(second)
	}))

	var all []*remotedomain.RemoteInstance
	require.NoError(t, tp.Read(func(r *appremote.TransactableRepositories) error {
		var err error
		all, err = r.RemoteInstances.GetAll()
		return err
	}))

	ids := make(map[string]struct{})
	for _, instance := range all {
		ids[instance.Id().String()] = struct{}{}
	}
	require.Len(t, ids, 2)
	require.Contains(t, ids, first.Id().String())
	require.Contains(t, ids, second.Id().String())
}

func TestGetByLocalPairingTokenFindsMatchingInstance(t *testing.T) {
	tp, _, cleanup := fixture.RemoteTransactionProvider(t)
	defer cleanup()

	wanted, token := newTestInstanceWithPairingToken(t)
	other := newTestInstance(t)

	require.NoError(t, tp.Write(func(r *appremote.TransactableRepositories) error {
		if err := r.RemoteInstances.Save(wanted); err != nil {
			return err
		}
		return r.RemoteInstances.Save(other)
	}))

	require.NoError(t, tp.Read(func(r *appremote.TransactableRepositories) error {
		got, err := r.RemoteInstances.GetByLocalPairingTokenHash(token.Hash())
		require.NoError(t, err)
		require.Equal(t, wanted.Id().String(), got.Id().String())
		return nil
	}))
}

func TestGetByLocalPairingTokenReturnsNotFound(t *testing.T) {
	tp, _, cleanup := fixture.RemoteTransactionProvider(t)
	defer cleanup()

	require.NoError(t, tp.Write(func(r *appremote.TransactableRepositories) error {
		return r.RemoteInstances.Save(newTestInstance(t))
	}))

	token, err := remotedomain.NewPairingToken()
	require.NoError(t, err)

	require.NoError(t, tp.Read(func(r *appremote.TransactableRepositories) error {
		_, err := r.RemoteInstances.GetByLocalPairingTokenHash(token.Hash())
		require.ErrorIs(t, err, appremote.ErrNotFound)
		return nil
	}))
}

func TestGetByLocalAuthTokenFindsMatchingInstance(t *testing.T) {
	tp, _, cleanup := fixture.RemoteTransactionProvider(t)
	defer cleanup()

	wanted := newTestInstance(t)
	authToken, err := wanted.IssueLocalAuthToken()
	require.NoError(t, err)
	other := newTestInstance(t)

	require.NoError(t, tp.Write(func(r *appremote.TransactableRepositories) error {
		if err := r.RemoteInstances.Save(wanted); err != nil {
			return err
		}
		return r.RemoteInstances.Save(other)
	}))

	require.NoError(t, tp.Read(func(r *appremote.TransactableRepositories) error {
		got, err := r.RemoteInstances.GetByLocalAuthTokenHash(authToken.Hash())
		require.NoError(t, err)
		require.Equal(t, wanted.Id().String(), got.Id().String())
		return nil
	}))
}

func TestGetByLocalAuthTokenReturnsNotFound(t *testing.T) {
	tp, _, cleanup := fixture.RemoteTransactionProvider(t)
	defer cleanup()

	// Saved instance has no local auth token hash at all.
	require.NoError(t, tp.Write(func(r *appremote.TransactableRepositories) error {
		return r.RemoteInstances.Save(newTestInstance(t))
	}))

	token, err := remotedomain.NewAuthToken()
	require.NoError(t, err)

	require.NoError(t, tp.Read(func(r *appremote.TransactableRepositories) error {
		_, err := r.RemoteInstances.GetByLocalAuthTokenHash(token.Hash())
		require.ErrorIs(t, err, appremote.ErrNotFound)
		return nil
	}))
}

func newTestInstance(t *testing.T) *remotedomain.RemoteInstance {
	t.Helper()
	instance, _ := newTestInstanceWithPairingToken(t)
	return instance
}

func newTestInstanceWithPairingToken(t *testing.T) (*remotedomain.RemoteInstance, remotedomain.PairingToken) {
	t.Helper()

	id, err := remotedomain.NewRemoteInstanceID()
	require.NoError(t, err)

	address, err := remotedomain.NewRemoteInstanceAddress("https://example.com")
	require.NoError(t, err)

	token, err := remotedomain.NewPairingToken()
	require.NoError(t, err)

	return remotedomain.NewRemoteInstance(id, address, token.Hash()), token
}
