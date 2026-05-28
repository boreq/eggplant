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

var (
	usersBucket         = []byte("users")
	sessionTokensBucket = []byte("session_tokens")
)

type UserRepository struct {
	tx  *bolt.Tx
	log logging.Logger
}

func NewUserRepository(tx *bolt.Tx) (*UserRepository, error) {
	if tx.Writable() {
		if _, err := tx.CreateBucketIfNotExists(usersBucket); err != nil {
			return nil, errors.Wrap(err, "could not create the users bucket")
		}
		if _, err := tx.CreateBucketIfNotExists(sessionTokensBucket); err != nil {
			return nil, errors.Wrap(err, "could not create the session tokens bucket")
		}
	}

	return &UserRepository{
		tx:  tx,
		log: logging.New("UserRepository"),
	}, nil
}

func (r *UserRepository) Count() (int, error) {
	b := r.tx.Bucket(usersBucket)
	if b == nil {
		return 0, nil
	}
	return b.Stats().KeyN, nil
}

func (r *UserRepository) List() ([]authdomain.User, error) {
	b := r.tx.Bucket(usersBucket)
	if b == nil {
		return nil, nil
	}

	c := b.Cursor()

	var users []authdomain.User
	for k, v := c.First(); k != nil; k, v = c.Next() {
		u, err := decodeUser(v)
		if err != nil {
			return nil, errors.Wrap(err, "could not build the user")
		}
		users = append(users, u)
	}

	return users, nil
}

func (r *UserRepository) Remove(username authdomain.Username) error {
	users := r.tx.Bucket(usersBucket)
	if users == nil {
		return errors.New("users bucket does not exist")
	}

	if existing := users.Get([]byte(username.String())); existing != nil {
		u, err := decodeUser(existing)
		if err != nil {
			return errors.Wrap(err, "could not decode the existing user")
		}
		if err := r.removeSessionTokens(u.Sessions()); err != nil {
			return errors.Wrap(err, "could not remove session tokens")
		}
	}

	return users.Delete([]byte(username.String()))
}

func (r *UserRepository) Get(username authdomain.Username) (*authdomain.User, error) {
	b := r.tx.Bucket(usersBucket)
	if b == nil {
		return nil, errors.Wrap(auth.ErrNotFound, "users bucket does not exist")
	}
	j := b.Get([]byte(username.String()))
	if j == nil {
		return nil, auth.ErrNotFound
	}

	u, err := decodeUser(j)
	if err != nil {
		return nil, errors.Wrap(err, "could not build the user")
	}

	return &u, nil
}

func (r *UserRepository) GetByToken(token authdomain.AccessToken) (*authdomain.User, error) {
	index := r.tx.Bucket(sessionTokensBucket)
	if index == nil {
		return nil, errors.Wrap(auth.ErrNotFound, "session tokens bucket does not exist")
	}
	v := index.Get([]byte(token.String()))
	if v == nil {
		return nil, auth.ErrNotFound
	}
	username, err := authdomain.NewUsernameFromString(string(v))
	if err != nil {
		return nil, errors.Wrap(err, "invalid username in index")
	}
	return r.Get(username)
}

func (r *UserRepository) Put(user authdomain.User) error {
	users := r.tx.Bucket(usersBucket)
	if users == nil {
		return errors.New("users bucket does not exist")
	}

	var priorSessions []authdomain.Session
	if existing := users.Get([]byte(user.Username().String())); existing != nil {
		prior, err := decodeUser(existing)
		if err != nil {
			return errors.Wrap(err, "could not decode the existing user")
		}
		priorSessions = prior.Sessions()
	}

	dto := userToDTO(user)
	dto.Sessions = removeOldSessions(dto.Sessions)

	j, err := json.Marshal(dto)
	if err != nil {
		return errors.Wrap(err, "marshaling to json failed")
	}

	if err := users.Put([]byte(dto.Username), j); err != nil {
		return errors.Wrap(err, "could not put the user")
	}

	return r.syncSessionTokens(user.Username(), priorSessions, dto.Sessions)
}

func (r *UserRepository) syncSessionTokens(username authdomain.Username, prior []authdomain.Session, current []sessionDTO) error {
	index := r.tx.Bucket(sessionTokensBucket)
	if index == nil {
		return errors.New("session tokens bucket does not exist")
	}

	keep := make(map[string]struct{}, len(current))
	for _, s := range current {
		keep[s.Token] = struct{}{}
	}

	for _, s := range prior {
		if _, ok := keep[s.Token().String()]; ok {
			continue
		}
		if err := index.Delete([]byte(s.Token().String())); err != nil {
			return errors.Wrap(err, "could not delete a stale session token")
		}
	}

	for tokenStr := range keep {
		if err := index.Put([]byte(tokenStr), []byte(username.String())); err != nil {
			return errors.Wrap(err, "could not put a session token")
		}
	}

	return nil
}

func (r *UserRepository) removeSessionTokens(sessions []authdomain.Session) error {
	index := r.tx.Bucket(sessionTokensBucket)
	if index == nil {
		return errors.New("session tokens bucket does not exist")
	}
	for _, s := range sessions {
		if err := index.Delete([]byte(s.Token().String())); err != nil {
			return errors.Wrap(err, "could not delete a session token")
		}
	}
	return nil
}

func decodeUser(j []byte) (authdomain.User, error) {
	var dto userDTO
	if err := json.Unmarshal(j, &dto); err != nil {
		return authdomain.User{}, errors.Wrap(err, "json unmarshal failed")
	}
	return userFromDTO(dto)
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
	}
	if c := u.Created(); c != nil {
		dto.Created = *c
	}
	if ls := u.LastSeen(); ls != nil {
		dto.LastSeen = *ls
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
	var created *time.Time
	if !dto.Created.IsZero() {
		created = &dto.Created
	}
	var lastSeen *time.Time
	if !dto.LastSeen.IsZero() {
		lastSeen = &dto.LastSeen
	}
	return authdomain.NewUserFromDatabase(
		username,
		password,
		dto.Administrator,
		created,
		lastSeen,
		sessions,
	), nil
}
