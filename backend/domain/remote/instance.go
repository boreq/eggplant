package remote

import (
	"time"

	"github.com/boreq/errors"
)

type RemoteInstance struct {
	id      RemoteInstanceID
	address RemoteInstanceAddress

	localPairingTokenHash *HashedPairingToken
	localAuthTokenHash    *HashedAuthToken

	remotePairingToken *PairingToken
	remoteAuthToken    *AuthToken

	lastHealthcheck *Healthcheck

	EventRecorder
}

func NewRemoteInstance(id RemoteInstanceID, address RemoteInstanceAddress, localPairingTokenHash HashedPairingToken) *RemoteInstance {
	return &RemoteInstance{
		id:      id,
		address: address,

		localPairingTokenHash: &localPairingTokenHash,
		localAuthTokenHash:    nil,

		remotePairingToken: nil,
		remoteAuthToken:    nil,

		lastHealthcheck: nil,
	}
}

func NewRemoteInstanceFromDatabase(
	id RemoteInstanceID,
	address RemoteInstanceAddress,
	localPairingTokenHash *HashedPairingToken,
	localAuthTokenHash *HashedAuthToken,
	remotePairingToken *PairingToken,
	remoteAuthToken *AuthToken,
	lastHealthcheck *Healthcheck,
) *RemoteInstance {
	return &RemoteInstance{
		id:                    id,
		address:               address,
		localPairingTokenHash: localPairingTokenHash,
		localAuthTokenHash:    localAuthTokenHash,
		remotePairingToken:    remotePairingToken,
		remoteAuthToken:       remoteAuthToken,
		lastHealthcheck:       lastHealthcheck,
	}
}

func (r *RemoteInstance) SetRemotePairingToken(token PairingToken) error {
	if r.remotePairingToken != nil && r.remotePairingToken.Equal(token) {
		return nil
	}

	if r.localAuthTokenHash != nil {
		return errors.New("the local auth token has already been issued for this remote instance")
	}

	r.remotePairingToken = &token
	r.RecordEvent(RemotePairingTokenSet{RemoteInstanceID: r.id})
	return nil
}

func (r *RemoteInstance) SetRemoteAuthToken(localPairingToken PairingToken, token AuthToken) error {
	if r.remoteAuthToken != nil && r.remoteAuthToken.Equal(token) {
		return nil
	}

	if r.localPairingTokenHash == nil {
		return errors.New("the local pairing token hash is not set")
	}

	if !r.localPairingTokenHash.Equal(localPairingToken.Hash()) {
		return errors.New("the local pairing token does not match")
	}

	wasPaired := r.isPaired()
	r.remoteAuthToken = &token
	r.localPairingTokenHash = nil
	r.recordPairedIfJustCompleted(wasPaired)
	return nil
}

func (r *RemoteInstance) Id() RemoteInstanceID {
	return r.id
}

func (r *RemoteInstance) Address() RemoteInstanceAddress {
	return r.address
}

func (r *RemoteInstance) LocalPairingTokenHash() (HashedPairingToken, bool) {
	if r.localPairingTokenHash == nil {
		return HashedPairingToken{}, false
	}
	return *r.localPairingTokenHash, true
}

func (r *RemoteInstance) RemotePairingToken() (PairingToken, bool) {
	if r.remotePairingToken == nil {
		return PairingToken{}, false
	}
	return *r.remotePairingToken, true
}

func (r *RemoteInstance) IssueLocalAuthToken() (AuthToken, error) {
	token, err := NewAuthToken()
	if err != nil {
		return AuthToken{}, errors.Wrap(err, "could not generate the local auth token")
	}
	wasPaired := r.isPaired()
	hash := token.Hash()
	r.localAuthTokenHash = &hash
	r.recordPairedIfJustCompleted(wasPaired)
	return token, nil
}

func (r *RemoteInstance) isPaired() bool {
	return r.remoteAuthToken != nil && r.localAuthTokenHash != nil
}

func (r *RemoteInstance) recordPairedIfJustCompleted(wasPaired bool) {
	if !wasPaired && r.isPaired() {
		r.RecordEvent(RemotePaired{RemoteInstanceID: r.id})
	}
}

func (r *RemoteInstance) LocalAuthTokenHash() (HashedAuthToken, bool) {
	if r.localAuthTokenHash == nil {
		return HashedAuthToken{}, false
	}
	return *r.localAuthTokenHash, true
}

func (r *RemoteInstance) RemoteAuthToken() (AuthToken, bool) {
	if r.remoteAuthToken == nil {
		return AuthToken{}, false
	}
	return *r.remoteAuthToken, true
}

func (r *RemoteInstance) CanBeQueried() bool {
	return r.remoteAuthToken != nil
}

func (r *RemoteInstance) Status() RemoteInstanceStatus {
	if !r.isPaired() {
		return RemoteInstanceStatusPairing
	}
	if r.lastHealthcheck != nil && r.lastHealthcheck.Status() == HealthcheckStatusAlive {
		return RemoteInstanceStatusHealthy
	}
	return RemoteInstanceStatusDead
}

func (r *RemoteInstance) RecordHealthcheck(status HealthcheckStatus, at time.Time) error {
	healthcheck, err := NewHealthcheck(status, at)
	if err != nil {
		return errors.Wrap(err, "could not create the healthcheck")
	}
	r.lastHealthcheck = &healthcheck
	return nil
}

func (r *RemoteInstance) LastHealthcheck() (Healthcheck, bool) {
	if r.lastHealthcheck == nil {
		return Healthcheck{}, false
	}
	return *r.lastHealthcheck, true
}
