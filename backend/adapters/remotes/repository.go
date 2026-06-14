package remotes

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"

	"github.com/boreq/errors"
	bolt "go.etcd.io/bbolt"
)

var topBucket = []byte("user_remotes")

type RemoteInstance struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Repository struct {
	db *bolt.DB
}

func NewRepository(db *bolt.DB) (*Repository, error) {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(topBucket)
		return err
	})
	if err != nil {
		return nil, errors.Wrap(err, "could not create remotes bucket")
	}
	return &Repository{db: db}, nil
}

func (r *Repository) List(username string) ([]RemoteInstance, error) {
	var result []RemoteInstance
	err := r.db.View(func(tx *bolt.Tx) error {
		top := tx.Bucket(topBucket)
		if top == nil {
			return nil
		}
		ub := top.Bucket([]byte(username))
		if ub == nil {
			return nil
		}
		c := ub.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var inst RemoteInstance
			if err := json.Unmarshal(v, &inst); err != nil {
				return errors.Wrap(err, "could not unmarshal remote instance")
			}
			result = append(result, inst)
		}
		return nil
	})
	return result, err
}

func (r *Repository) Get(username, id string) (*RemoteInstance, error) {
	var result *RemoteInstance
	err := r.db.View(func(tx *bolt.Tx) error {
		top := tx.Bucket(topBucket)
		if top == nil {
			return nil
		}
		ub := top.Bucket([]byte(username))
		if ub == nil {
			return nil
		}
		v := ub.Get([]byte(id))
		if v == nil {
			return nil
		}
		var inst RemoteInstance
		if err := json.Unmarshal(v, &inst); err != nil {
			return errors.Wrap(err, "could not unmarshal remote instance")
		}
		result = &inst
		return nil
	})
	return result, err
}

func (r *Repository) Put(username string, inst RemoteInstance) (RemoteInstance, error) {
	if inst.ID == "" {
		inst.ID = generateID()
	}
	err := r.db.Update(func(tx *bolt.Tx) error {
		top, err := tx.CreateBucketIfNotExists(topBucket)
		if err != nil {
			return errors.Wrap(err, "could not open top bucket")
		}
		ub, err := top.CreateBucketIfNotExists([]byte(username))
		if err != nil {
			return errors.Wrap(err, "could not create user bucket")
		}
		j, err := json.Marshal(inst)
		if err != nil {
			return errors.Wrap(err, "could not marshal remote instance")
		}
		return ub.Put([]byte(inst.ID), j)
	})
	return inst, err
}

func (r *Repository) Delete(username, id string) error {
	return r.db.Update(func(tx *bolt.Tx) error {
		top := tx.Bucket(topBucket)
		if top == nil {
			return nil
		}
		ub := top.Bucket([]byte(username))
		if ub == nil {
			return nil
		}
		return ub.Delete([]byte(id))
	})
}

func generateID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
