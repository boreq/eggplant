package remote

import (
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/boreq/eggplant/application/remote"
	"github.com/boreq/eggplant/domain/crockford"
	remotedomain "github.com/boreq/eggplant/domain/remote"
	"github.com/boreq/errors"
	bolt "go.etcd.io/bbolt"
)

var remoteInstancesBucket = []byte("remote_instances")

type RemoteInstanceRepository struct {
	tx *bolt.Tx
}

func NewRemoteInstanceRepository(tx *bolt.Tx) (*RemoteInstanceRepository, error) {
	if tx.Writable() {
		if _, err := tx.CreateBucketIfNotExists(remoteInstancesBucket); err != nil {
			return nil, errors.Wrap(err, "could not create a bucket")
		}
	}
	return &RemoteInstanceRepository{tx: tx}, nil
}

func (r *RemoteInstanceRepository) Save(instance *remotedomain.RemoteInstance) error {
	dto := toRemoteInstanceDTO(instance)

	j, err := json.Marshal(dto)
	if err != nil {
		return errors.Wrap(err, "marshaling to json failed")
	}

	b := r.tx.Bucket(remoteInstancesBucket)
	if b == nil {
		return errors.New("bucket does not exist")
	}

	return b.Put([]byte(dto.ID), j)
}

func (r *RemoteInstanceRepository) GetByID(id remotedomain.RemoteInstanceID) (*remotedomain.RemoteInstance, error) {
	b := r.tx.Bucket(remoteInstancesBucket)
	if b == nil {
		return nil, errors.Wrap(remote.ErrNotFound, "bucket does not exist")
	}

	j := b.Get([]byte(id.String()))
	if j == nil {
		return nil, remote.ErrNotFound
	}

	return unmarshalRemoteInstance(j)
}

func (r *RemoteInstanceRepository) GetAll() ([]*remotedomain.RemoteInstance, error) {
	b := r.tx.Bucket(remoteInstancesBucket)
	if b == nil {
		return nil, nil
	}

	var instances []*remotedomain.RemoteInstance
	if err := b.ForEach(func(key, value []byte) error {
		instance, err := unmarshalRemoteInstance(value)
		if err != nil {
			return errors.Wrap(err, "could not unmarshal the instance")
		}
		instances = append(instances, instance)
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "for each failed")
	}

	return instances, nil
}

func (r *RemoteInstanceRepository) GetByLocalPairingTokenHash(hash remotedomain.HashedPairingToken) (*remotedomain.RemoteInstance, error) {
	encoded := hex.EncodeToString(hash.Bytes())
	return r.findBy(func(dto remoteInstanceDTO) bool {
		return dto.LocalPairingTokenHash != "" && dto.LocalPairingTokenHash == encoded
	})
}

func (r *RemoteInstanceRepository) GetByLocalAuthTokenHash(hash remotedomain.HashedAuthToken) (*remotedomain.RemoteInstance, error) {
	encoded := hex.EncodeToString(hash.Bytes())
	return r.findBy(func(dto remoteInstanceDTO) bool {
		return dto.LocalAuthTokenHash != "" && dto.LocalAuthTokenHash == encoded
	})
}

func (r *RemoteInstanceRepository) findBy(predicate func(remoteInstanceDTO) bool) (*remotedomain.RemoteInstance, error) {
	b := r.tx.Bucket(remoteInstancesBucket)
	if b == nil {
		return nil, errors.Wrap(remote.ErrNotFound, "bucket does not exist")
	}

	var found []byte
	if err := b.ForEach(func(key, value []byte) error {
		var dto remoteInstanceDTO
		if err := json.Unmarshal(value, &dto); err != nil {
			return errors.Wrap(err, "json unmarshal failed")
		}
		if predicate(dto) {
			found = value
		}
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "for each failed")
	}

	if found == nil {
		return nil, remote.ErrNotFound
	}

	return unmarshalRemoteInstance(found)
}

type remoteInstanceDTO struct {
	ID                    string `json:"id"`
	Address               string `json:"address"`
	LocalPairingTokenHash string `json:"localPairingTokenHash,omitempty"`
	LocalAuthTokenHash    string `json:"localAuthTokenHash,omitempty"`
	RemotePairingToken    string `json:"remotePairingToken,omitempty"`
	RemoteAuthToken       string `json:"remoteAuthToken,omitempty"`
	LastHealthcheckStatus string `json:"lastHealthcheckStatus,omitempty"`
	LastHealthcheckTime   string `json:"lastHealthcheckTime,omitempty"`
}

func toRemoteInstanceDTO(instance *remotedomain.RemoteInstance) remoteInstanceDTO {
	dto := remoteInstanceDTO{
		ID:      instance.Id().String(),
		Address: instance.Address().String(),
	}
	if h, ok := instance.LocalAuthTokenHash(); ok {
		dto.LocalAuthTokenHash = hex.EncodeToString(h.Bytes())
	}
	if h, ok := instance.LocalPairingTokenHash(); ok {
		dto.LocalPairingTokenHash = hex.EncodeToString(h.Bytes())
	}
	if t, ok := instance.RemotePairingToken(); ok {
		dto.RemotePairingToken = crockford.Encode(t.Bytes())
	}
	if t, ok := instance.RemoteAuthToken(); ok {
		dto.RemoteAuthToken = crockford.Encode(t.Bytes())
	}
	if h, ok := instance.LastHealthcheck(); ok {
		dto.LastHealthcheckStatus = h.Status().String()
		dto.LastHealthcheckTime = h.At().UTC().Format(time.RFC3339Nano)
	}
	return dto
}

func unmarshalRemoteInstance(j []byte) (*remotedomain.RemoteInstance, error) {
	var dto remoteInstanceDTO
	if err := json.Unmarshal(j, &dto); err != nil {
		return nil, errors.Wrap(err, "json unmarshal failed")
	}
	return remoteInstanceFromDTO(dto)
}

func remoteInstanceFromDTO(dto remoteInstanceDTO) (*remotedomain.RemoteInstance, error) {
	id, err := remotedomain.NewRemoteInstanceIDFromString(dto.ID)
	if err != nil {
		return nil, errors.Wrap(err, "invalid id")
	}

	address, err := remotedomain.NewRemoteInstanceAddress(dto.Address)
	if err != nil {
		return nil, errors.Wrap(err, "invalid address")
	}

	localPairingTokenHash, err := optionalHashedPairingTokenFromDTO(dto.LocalPairingTokenHash)
	if err != nil {
		return nil, errors.Wrap(err, "invalid local pairing token hash")
	}

	remotePairingToken, err := optionalPairingTokenFromDTO(dto.RemotePairingToken)
	if err != nil {
		return nil, errors.Wrap(err, "invalid remote pairing token")
	}

	remoteAuthToken, err := optionalAuthTokenFromDTO(dto.RemoteAuthToken)
	if err != nil {
		return nil, errors.Wrap(err, "invalid remote auth token")
	}

	localAuthTokenHash, err := optionalHashedAuthTokenFromDTO(dto.LocalAuthTokenHash)
	if err != nil {
		return nil, errors.Wrap(err, "invalid local auth token hash")
	}

	lastHealthcheck, err := optionalHealthcheckFromDTO(dto.LastHealthcheckStatus, dto.LastHealthcheckTime)
	if err != nil {
		return nil, errors.Wrap(err, "invalid last healthcheck")
	}

	return remotedomain.NewRemoteInstanceFromDatabase(id, address, localPairingTokenHash, localAuthTokenHash, remotePairingToken, remoteAuthToken, lastHealthcheck), nil
}

func optionalHealthcheckFromDTO(status, at string) (*remotedomain.Healthcheck, error) {
	if status == "" && at == "" {
		return nil, nil
	}
	s, err := remotedomain.NewHealthcheckStatusFromString(status)
	if err != nil {
		return nil, errors.Wrap(err, "invalid status")
	}
	t, err := time.Parse(time.RFC3339Nano, at)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse the time")
	}
	healthcheck, err := remotedomain.NewHealthcheck(s, t)
	if err != nil {
		return nil, errors.Wrap(err, "could not create the healthcheck")
	}
	return &healthcheck, nil
}

func optionalHashedAuthTokenFromDTO(s string) (*remotedomain.HashedAuthToken, error) {
	if s == "" {
		return nil, nil
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode the hash")
	}
	hash, err := remotedomain.NewHashedAuthTokenFromBytes(b)
	if err != nil {
		return nil, errors.Wrap(err, "could not create the hashed auth token")
	}
	return &hash, nil
}

func optionalHashedPairingTokenFromDTO(s string) (*remotedomain.HashedPairingToken, error) {
	if s == "" {
		return nil, nil
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode the hash")
	}
	hash, err := remotedomain.NewHashedPairingTokenFromBytes(b)
	if err != nil {
		return nil, errors.Wrap(err, "could not create the hashed pairing token")
	}
	return &hash, nil
}

func optionalPairingTokenFromDTO(s string) (*remotedomain.PairingToken, error) {
	if s == "" {
		return nil, nil
	}
	b, err := crockford.Decode(s)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode the token")
	}
	token, err := remotedomain.NewPairingTokenFromBytes(b)
	if err != nil {
		return nil, errors.Wrap(err, "could not create the pairing token")
	}
	return &token, nil
}

func optionalAuthTokenFromDTO(s string) (*remotedomain.AuthToken, error) {
	if s == "" {
		return nil, nil
	}
	b, err := crockford.Decode(s)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode the token")
	}
	token, err := remotedomain.NewAuthTokenFromBytes(b)
	if err != nil {
		return nil, errors.Wrap(err, "could not create the auth token")
	}
	return &token, nil
}
