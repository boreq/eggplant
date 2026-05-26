package auth

import (
	"encoding/json"
	"time"

	"github.com/boreq/eggplant/application/auth"
	authdomain "github.com/boreq/eggplant/domain/auth"
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

func (r *UserRepository) List() ([]authdomain.User, error) {
	b := r.tx.Bucket(r.bucket)
	if b == nil {
		return nil, nil
	}

	c := b.Cursor()

	var users []authdomain.User
	for k, v := c.First(); k != nil; k, v = c.Next() {
		var dto userDTO
		if err := json.Unmarshal(v, &dto); err != nil {
			return nil, errors.Wrap(err, "json unmarshal failed")
		}
		u, err := userFromDTO(dto)
		if err != nil {
			return nil, errors.Wrap(err, "could not build the user")
		}
		users = append(users, u)
	}

	return users, nil
}

func (r *UserRepository) Remove(username authdomain.Username) error {
	b := r.tx.Bucket(r.bucket)
	if b == nil {
		return errors.New("bucket does not exist")
	}
	return b.Delete([]byte(username.String()))
}

func (r *UserRepository) Get(username authdomain.Username) (*authdomain.User, error) {
	b := r.tx.Bucket(r.bucket)
	if b == nil {
		return nil, errors.Wrap(auth.ErrNotFound, "bucket does not exist")
	}
	j := b.Get([]byte(username.String()))
	if j == nil {
		return nil, auth.ErrNotFound
	}

	var dto userDTO
	if err := json.Unmarshal(j, &dto); err != nil {
		return nil, errors.Wrap(err, "json unmarshal failed")
	}

	u, err := userFromDTO(dto)
	if err != nil {
		return nil, errors.Wrap(err, "could not build the user")
	}

	return &u, nil
}

func (r *UserRepository) Put(user authdomain.User) error {
	dto := userToDTO(user)
	dto.Sessions = removeOldSessions(dto.Sessions)

	j, err := json.Marshal(dto)
	if err != nil {
		return errors.Wrap(err, "marshaling to json failed")
	}

	b := r.tx.Bucket(r.bucket)
	if b == nil {
		return errors.New("bucket does not exist")
	}
	return b.Put([]byte(dto.Username), j)
}

func removeOldSessions(sessions []sessionDTO) []sessionDTO {
	var result []sessionDTO
	for _, session := range sessions {
		if session.LastSeen.Add(365 * 24 * time.Hour).After(time.Now()) {
			result = append(result, session)
		}
	}
	return result
}

type userDTO struct {
	Username      string       `json:"username"`
	Password      []byte       `json:"password"`
	Administrator bool         `json:"administrator"`
	Created       time.Time    `json:"created"`
	LastSeen      time.Time    `json:"lastSeen"`
	Sessions      []sessionDTO `json:"sessions"`
}

type sessionDTO struct {
	Token    string    `json:"token"`
	LastSeen time.Time `json:"lastSeen"`
}

func userToDTO(u authdomain.User) userDTO {
	dto := userDTO{
		Username:      u.Username().String(),
		Password:      u.Password().Bytes(),
		Administrator: u.Administrator(),
		Created:       u.Created(),
		LastSeen:      u.LastSeen(),
	}
	for _, s := range u.Sessions() {
		dto.Sessions = append(dto.Sessions, sessionDTO{
			Token:    s.Token().String(),
			LastSeen: s.LastSeen(),
		})
	}
	return dto
}

func userFromDTO(dto userDTO) (authdomain.User, error) {
	username, err := authdomain.NewUsernameFromString(dto.Username)
	if err != nil {
		return authdomain.User{}, errors.Wrap(err, "invalid username")
	}
	password, err := authdomain.NewPasswordHash(dto.Password)
	if err != nil {
		return authdomain.User{}, errors.Wrap(err, "invalid password hash")
	}
	var sessions []authdomain.Session
	for _, s := range dto.Sessions {
		token, err := authdomain.NewAccessTokenFromString(s.Token)
		if err != nil {
			return authdomain.User{}, errors.Wrap(err, "invalid session token")
		}
		sessions = append(sessions, authdomain.NewSession(token, s.LastSeen))
	}
	return authdomain.NewUser(
		username,
		password,
		dto.Administrator,
		dto.Created,
		dto.LastSeen,
		sessions,
	), nil
}
