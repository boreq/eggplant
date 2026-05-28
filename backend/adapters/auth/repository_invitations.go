package auth

import (
	"encoding/json"
	"time"

	"github.com/boreq/eggplant/application/auth"
	authdomain "github.com/boreq/eggplant/domain/auth"
	"github.com/boreq/eggplant/internal/logging"
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

func (r *InvitationRepository) Put(invitation authdomain.Invitation) error {
	dto := invitationToDTO(invitation)

	j, err := json.Marshal(dto)
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

	return b.Put([]byte(dto.Token), j)
}

func (r *InvitationRepository) Get(token authdomain.InvitationToken) (*authdomain.Invitation, error) {
	b := r.tx.Bucket(r.bucket)
	if b == nil {
		return nil, errors.Wrap(auth.ErrNotFound, "bucket does not exist")
	}
	j := b.Get([]byte(token.String()))
	if j == nil {
		return nil, auth.ErrNotFound
	}

	var dto invitationDTO
	if err := json.Unmarshal(j, &dto); err != nil {
		return nil, errors.Wrap(err, "json unmarshal failed")
	}

	if time.Now().After(dto.Created.Add(maxInvitationAge)) {
		return nil, auth.ErrNotFound
	}

	invitation, err := invitationFromDTO(dto)
	if err != nil {
		return nil, errors.Wrap(err, "could not build the invitation")
	}

	return &invitation, nil
}

func (r *InvitationRepository) Remove(token authdomain.InvitationToken) error {
	b := r.tx.Bucket(r.bucket)
	if b == nil {
		return errors.New("bucket does not exist")
	}
	return b.Delete([]byte(token.String()))
}

func (r *InvitationRepository) removeOldInvitations(b *bolt.Bucket) error {
	var keysToRemove [][]byte

	if err := b.ForEach(func(key, value []byte) error {
		var dto invitationDTO
		if err := json.Unmarshal(value, &dto); err != nil {
			return errors.Wrap(err, "json unmarshal failed")
		}

		if time.Now().After(dto.Created.Add(maxInvitationAge)) {
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

type invitationDTO struct {
	Token   string    `json:"invitation"`
	Created time.Time `json:"created"`
}

func invitationToDTO(i authdomain.Invitation) invitationDTO {
	return invitationDTO{
		Token:   i.Token().String(),
		Created: i.Created(),
	}
}

func invitationFromDTO(dto invitationDTO) (authdomain.Invitation, error) {
	token, err := authdomain.NewInvitationTokenFromString(dto.Token)
	if err != nil {
		return authdomain.Invitation{}, errors.Wrap(err, "invalid invitation token")
	}
	return authdomain.NewInvitation(token, dto.Created), nil
}
