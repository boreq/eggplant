package auth

import (
	"github.com/boreq/eggplant/application/auth"
	authdomain "github.com/boreq/eggplant/domain/auth"
	"github.com/boreq/errors"
	bolt "go.etcd.io/bbolt"
)

type SessionTokenRepository struct {
	tx     *bolt.Tx
	bucket []byte
}

func NewSessionTokenRepository(tx *bolt.Tx) (*SessionTokenRepository, error) {
	bucket := []byte("session_tokens")

	if tx.Writable() {
		if _, err := tx.CreateBucketIfNotExists(bucket); err != nil {
			return nil, errors.Wrap(err, "could not create a bucket")
		}
	}

	return &SessionTokenRepository{
		tx:     tx,
		bucket: bucket,
	}, nil
}

func (r *SessionTokenRepository) Put(token authdomain.AccessToken, username authdomain.Username) error {
	b := r.tx.Bucket(r.bucket)
	if b == nil {
		return errors.New("bucket does not exist")
	}
	return b.Put([]byte(token.String()), []byte(username.String()))
}

func (r *SessionTokenRepository) Get(token authdomain.AccessToken) (authdomain.Username, error) {
	b := r.tx.Bucket(r.bucket)
	if b == nil {
		return authdomain.Username{}, errors.Wrap(auth.ErrNotFound, "bucket does not exist")
	}
	v := b.Get([]byte(token.String()))
	if v == nil {
		return authdomain.Username{}, auth.ErrNotFound
	}
	username, err := authdomain.NewUsernameFromString(string(v))
	if err != nil {
		return authdomain.Username{}, errors.Wrap(err, "invalid username in index")
	}
	return username, nil
}

func (r *SessionTokenRepository) Remove(token authdomain.AccessToken) error {
	b := r.tx.Bucket(r.bucket)
	if b == nil {
		return errors.New("bucket does not exist")
	}
	return b.Delete([]byte(token.String()))
}
