package auth

import (
	"encoding/json"

	"github.com/boreq/eggplant/application/auth"
	"github.com/boreq/eggplant/logging"
	"github.com/boreq/errors"
	bolt "go.etcd.io/bbolt"
)

type InvitationRepository struct {
	tx     *bolt.Tx
	bucket []byte
	log    logging.Logger
}

func NewInvitationRepository(tx *bolt.Tx) (*InvitationRepository, error) {
	bucket := []byte("invitations")

	if tx.Writable() {
		if _, err := tx.CreateBucketIfNotExists(bucket); err != nil {
			return nil, errors.Wrap(err, "could not create a bucket")
		}
	}

	return &InvitationRepository{
		tx:     tx,
		bucket: bucket,
		log:    logging.New("InvitationRepository"),
	}, nil
}

func (r *InvitationRepository) Put(invitation auth.Invitation) error {
	j, err := json.Marshal(invitation)
	if err != nil {
		return errors.Wrap(err, "marshaling to json failed")
	}

	b := r.tx.Bucket(r.bucket)
	if b == nil {
		return errors.New("bucket does not exist")
	}
	return b.Put([]byte(invitation.Token), j)
}

func (r *InvitationRepository) Get(token auth.InvitationToken) (*auth.Invitation, error) {
	b := r.tx.Bucket(r.bucket)
	if b == nil {
		return nil, errors.Wrap(auth.ErrNotFound, "bucket does not exist")
	}
	j := b.Get([]byte(token))
	if j == nil {
		return nil, auth.ErrNotFound
	}

	invitation := &auth.Invitation{}
	if err := json.Unmarshal(j, invitation); err != nil {
		return nil, errors.Wrap(err, "json unmarshal failed")
	}

	return invitation, nil
}

func (r *InvitationRepository) Remove(token auth.InvitationToken) error {
	b := r.tx.Bucket(r.bucket)
	if b == nil {
		return errors.New("bucket does not exist")
	}
	return b.Delete([]byte(token))
}
