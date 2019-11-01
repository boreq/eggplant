package auth

import (
	"encoding/json"
	"time"

	"github.com/boreq/eggplant/errors"
	"github.com/boreq/eggplant/logging"
	"github.com/boreq/eggplant/pkg/service/application/auth"
	bolt "go.etcd.io/bbolt"
)

type PasswordHash []byte

type PasswordHasher interface {
	Hash(password string) (PasswordHash, error)
	Compare(hashedPassword PasswordHash, password string) error
}

type AccessTokenGenerator interface {
	Generate(username string) (auth.AccessToken, error)
	GetUsername(token auth.AccessToken) (string, error)
}

type user struct {
	Username      string       `json:"username"`
	Password      PasswordHash `json:"password"`
	Administrator bool         `json:"administrator"`
	Sessions      []session    `json:"sessions"`
}

type invitation struct {
	Token   auth.InvitationToken `json:"token"`
	Created time.Time            `json:"created"`
}

type session struct {
	Token    auth.AccessToken `json:"token"`
	LastSeen time.Time        `json:"lastSeen"`
}

type UserRepository struct {
	db                   *bolt.DB
	passwordHasher       PasswordHasher
	accessTokenGenerator AccessTokenGenerator
	usersBucket          []byte
	invitationsBucket    []byte
	log                  logging.Logger
}

func NewUserRepository(
	db *bolt.DB,
	passwordHasher PasswordHasher,
	accessTokenGenerator AccessTokenGenerator,
) (*UserRepository, error) {
	usersBucket := []byte("users")
	invitationsBucket := []byte("invitations")

	if err := db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(usersBucket); err != nil {
			return errors.Wrap(err, "could not create a users bucket")
		}
		if _, err := tx.CreateBucketIfNotExists(invitationsBucket); err != nil {
			return errors.Wrap(err, "could not create an invitations bucket")
		}
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "update failed")
	}

	return &UserRepository{
		passwordHasher:       passwordHasher,
		accessTokenGenerator: accessTokenGenerator,
		db:                   db,
		usersBucket:          usersBucket,
		invitationsBucket:    invitationsBucket,
		log:                  logging.New("userRepository"),
	}, nil
}

func (r *UserRepository) RegisterInitial(username, password string) error {
	if err := r.validate(username, password); err != nil {
		return errors.Wrap(err, "invalid parameters")
	}

	passwordHash, err := r.passwordHasher.Hash(password)
	if err != nil {
		return errors.Wrap(err, "hashing the password failed")
	}

	u := user{
		Username:      username,
		Password:      passwordHash,
		Administrator: true,
	}

	j, err := json.Marshal(u)
	if err != nil {
		return errors.Wrap(err, "marshaling to json failed")
	}

	return r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(r.usersBucket)
		if !bucketIsEmpty(b) {
			return errors.New("there are existing users")
		}
		return b.Put([]byte(u.Username), j)
	})
}

func (r *UserRepository) Login(username, password string) (auth.AccessToken, error) {
	if err := r.validate(username, password); err != nil {
		return "", auth.ErrUnauthorized
	}

	var token auth.AccessToken

	if err := r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(r.usersBucket)
		j := b.Get([]byte(username))
		if j == nil {
			return auth.ErrUnauthorized
		}

		var u user
		if err := json.Unmarshal(j, &u); err != nil {
			return errors.Wrap(err, "json unmarshal failed")
		}

		if err := r.passwordHasher.Compare(u.Password, password); err != nil {
			return auth.ErrUnauthorized
		}

		t, err := r.accessTokenGenerator.Generate(username)
		if err != nil {
			return errors.Wrap(err, "could not create an access token")
		}
		token = t

		s := session{
			Token: t,
		}

		u.Sessions = append(u.Sessions, s)

		j, err = json.Marshal(u)
		if err != nil {
			return errors.Wrap(err, "marshaling to json failed")
		}

		return b.Put([]byte(username), j)
	}); err != nil {
		return "", errors.Wrap(err, "transaction failed")
	}

	return token, nil

}

func (r *UserRepository) CheckAccessToken(token auth.AccessToken) (auth.User, error) {
	username, err := r.accessTokenGenerator.GetUsername(token)
	if err != nil {
		r.log.Warn("could not get the username", "err", err)
		return auth.User{}, auth.ErrUnauthorized
	}

	var foundUser user
	if err := r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(r.usersBucket)

		u, err := r.getUser(b, username)
		if err != nil {
			return errors.Wrap(err, "could not get the user")
		}

		if u == nil {
			r.log.Warn("user does't exist", "username", username)
			return auth.ErrUnauthorized
		}

		for i := range u.Sessions {
			if u.Sessions[i].Token == token {
				u.Sessions[i].LastSeen = time.Now()
				foundUser = *u
				return r.putUser(b, *u)
			}
		}

		return errors.New("invalid token")
	}); err != nil {
		return auth.User{}, errors.Wrap(err, "transaction failed")
	}

	return r.toUser(foundUser), nil
}

func (r *UserRepository) Logout(token auth.AccessToken) error {
	username, err := r.accessTokenGenerator.GetUsername(token)
	if err != nil {
		return errors.Wrap(err, "could not extract the username")
	}

	if err := r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(r.usersBucket)

		u, err := r.getUser(b, username)
		if err != nil {
			return errors.Wrap(err, "could not get the user")
		}

		if u == nil {
			return errors.New("user doesn't exist")
		}

		for i := range u.Sessions {
			if u.Sessions[i].Token == token {
				u.Sessions = append(u.Sessions[:i], u.Sessions[i+1:]...)
				return r.putUser(b, *u)
			}
		}

		return errors.New("session not found")
	}); err != nil {
		return errors.Wrap(err, "transaction failed")
	}

	return nil
}

func (r *UserRepository) validate(username, password string) error {
	if username == "" {
		return errors.New("username can't be empty")
	}

	if password == "" {
		return errors.New("password can't be empty")
	}

	return nil
}

func (r *UserRepository) Count() (int, error) {
	var count int
	if err := r.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(r.usersBucket)
		count = b.Stats().KeyN
		return nil
	}); err != nil {
		return 0, errors.Wrap(err, "view error")
	}
	return count, nil
}

func (r *UserRepository) List() ([]auth.User, error) {
	var users []auth.User
	if err := r.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(r.usersBucket)
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var u user
			if err := json.Unmarshal(v, &u); err != nil {
				return errors.Wrap(err, "json unmarshal failed")
			}
			users = append(users, r.toUser(u))
		}

		return nil
	}); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) CreateInvitation() (auth.InvitationToken, error) {
	s, err := generateCryptoString(256 / 8)
	if err != nil {
		return "", errors.Wrap(err, "could not create a token")
	}

	token := auth.InvitationToken(s)

	i := invitation{
		Token:   token,
		Created: time.Now(),
	}

	j, err := json.Marshal(i)
	if err != nil {
		return "", errors.Wrap(err, "marshaling to json failed")
	}

	if err := r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(r.invitationsBucket)
		return b.Put([]byte(token), j)
	}); err != nil {
		return "", errors.Wrap(err, "transaction failed")
	}

	return token, nil
}

func (r *UserRepository) getUser(b *bolt.Bucket, username string) (*user, error) {
	j := b.Get([]byte(username))
	if j == nil {
		return nil, nil
	}

	u := &user{}
	if err := json.Unmarshal(j, u); err != nil {
		return nil, errors.Wrap(err, "json unmarshal failed")
	}

	return u, nil
}

func (r *UserRepository) putUser(b *bolt.Bucket, u user) error {
	j, err := json.Marshal(u)
	if err != nil {
		return errors.Wrap(err, "marshaling to json failed")
	}

	return b.Put([]byte(u.Username), j)
}

func (r *UserRepository) toUser(u user) auth.User {
	return auth.User{
		Username:      u.Username,
		Administrator: u.Administrator,
	}
}

func bucketIsEmpty(b *bolt.Bucket) bool {
	key, value := b.Cursor().First()
	return key == nil && value == nil
}
