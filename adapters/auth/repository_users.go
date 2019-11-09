package auth

import (
	"encoding/json"

	"github.com/boreq/eggplant/application/auth"
	"github.com/boreq/eggplant/logging"
	"github.com/boreq/errors"
	bolt "go.etcd.io/bbolt"
)

type UserRepository struct {
	tx     *bolt.Tx
	bucket []byte
	log    logging.Logger
}

func NewUserRepository(tx *bolt.Tx) (*UserRepository, error) {
	bucket := []byte("users")

	if tx.Writable() {
		if _, err := tx.CreateBucketIfNotExists(bucket); err != nil {
			return nil, errors.Wrap(err, "could not create a bucket")
		}
	}

	return &UserRepository{
		tx:     tx,
		bucket: bucket,
		log:    logging.New("UserRepository"),
	}, nil
}

func (r *UserRepository) Count() (int, error) {
	b := r.tx.Bucket(r.bucket)
	if b == nil {
		return 0, nil
	}
	count := b.Stats().KeyN
	return count, nil
}

func (r *UserRepository) List() ([]auth.User, error) {
	b := r.tx.Bucket(r.bucket)
	if b == nil {
		return nil, nil
	}

	c := b.Cursor()

	var users []auth.User
	for k, v := c.First(); k != nil; k, v = c.Next() {
		var u auth.User
		if err := json.Unmarshal(v, &u); err != nil {
			return nil, errors.Wrap(err, "json unmarshal failed")
		}
		users = append(users, u)
	}

	return users, nil
}

func (r *UserRepository) Remove(username string) error {
	b := r.tx.Bucket(r.bucket)
	if b == nil {
		return errors.New("bucket does not exist")
	}
	return b.Delete([]byte(username))
}

func (r *UserRepository) Get(username string) (*auth.User, error) {
	b := r.tx.Bucket(r.bucket)
	if b == nil {
		return nil, errors.Wrap(auth.ErrNotFound, "bucket does not exist")
	}
	j := b.Get([]byte(username))
	if j == nil {
		return nil, auth.ErrNotFound
	}

	u := &auth.User{}
	if err := json.Unmarshal(j, u); err != nil {
		return nil, errors.Wrap(err, "json unmarshal failed")
	}

	return u, nil
}

func (r *UserRepository) Put(user auth.User) error {
	j, err := json.Marshal(user)
	if err != nil {
		return errors.Wrap(err, "marshaling to json failed")
	}

	b := r.tx.Bucket(r.bucket)
	if b == nil {
		return errors.New("bucket does not exist")
	}
	return b.Put([]byte(user.Username), j)
}
