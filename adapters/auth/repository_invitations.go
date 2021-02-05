package auth

import (
	"encoding/json"
	"time"

	"github.com/boreq/eggplant/application/auth"
	"github.com/boreq/eggplant/logging"
	"github.com/boreq/errors"
	bolt "go.etcd.io/bbolt"
)

const maxInvitationAge = 48 * time.Hour

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

	if err := r.removeOldInvitations(b); err != nil {
		return errors.Wrap(err, "could not remove old invitations")
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

	if time.Now().After(invitation.Created.Add(maxInvitationAge)) {
		return nil, auth.ErrNotFound
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

func (r *InvitationRepository) removeOldInvitations(b *bolt.Bucket) error {
	var keysToRemove [][]byte

	if err := b.ForEach(func(key, value []byte) error {
		invitation := &auth.Invitation{}
		if err := json.Unmarshal(value, invitation); err != nil {
			return errors.Wrap(err, "json unmarshal failed")
		}

		if time.Now().After(invitation.Created.Add(maxInvitationAge)) {
			keysToRemove = append(keysToRemove, nil)
		}

		return nil
	}); err != nil {
		return errors.Wrap(err, "for each failed")
	}

	for _, key := range keysToRemove {
		if err := b.Delete(key); err != nil {
			return errors.Wrap(err, "delete failed")
		}
	}

	return nil
}
