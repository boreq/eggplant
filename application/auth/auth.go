package auth

import (
	"errors"
	"fmt"
	"time"
)

var ErrUnauthorized = errors.New("unauthorized")
var ErrUsernameTaken = errors.New("username taken")
var ErrNotFound = errors.New("not found")

type CryptoStringGenerator interface {
	Generate(bytes int) (string, error)
}

type AccessTokenGenerator interface {
	Generate(username string) (AccessToken, error)
	GetUsername(token AccessToken) (string, error)
}

type PasswordHasher interface {
	Hash(password string) (PasswordHash, error)
	Compare(hashedPassword PasswordHash, password string) error
}

type UserRepository interface {
	// Put inserts the user into the repository. The previous entry with
	// this username is overwriten.
	Put(user User) error

	// Get returns the user with the provided username. If the user doesn't
	// exist ErrNotFound is returned.
	Get(username string) (*User, error)

	// Remove should remove a user.
	Remove(username string) error

	// List should return a list of all users.
	List() ([]User, error)

	// Count returns the number of users.
	Count() (int, error)
}

type InvitationRepository interface {
	// Put inserts the invitation into the repository. The previous entry
	// with this token is overwriten.
	Put(invitation Invitation) error

	// Get returns an invitation with the provided token, if the invitation
	// doesn't exist ErrNotFound is returned.
	Get(token InvitationToken) (*Invitation, error)

	// Remove removes an invitation. If the invitation doesn't exist this
	// function returns nil.
	Remove(token InvitationToken) error
}

type LastSeenUpdater interface {
	Update(username string, token AccessToken, t time.Time)
}

type AccessToken string

type InvitationToken string

type PasswordHash []byte

type User struct {
	Username      string       `json:"username"`
	Password      PasswordHash `json:"password"`
	Administrator bool         `json:"administrator"`
	Created       time.Time    `json:"created"`
	LastSeen      time.Time    `json:"lastSeen"`
	Sessions      []Session    `json:"sessions"`
}

type Session struct {
	Token    AccessToken `json:"token"`
	LastSeen time.Time   `json:"lastSeen"`
}

type ReadUser struct {
	Username      string        `json:"username"`
	Administrator bool          `json:"administrator"`
	Created       time.Time     `json:"created"`
	LastSeen      time.Time     `json:"lastSeen"`
	Sessions      []ReadSession `json:"sessions"`
}

type ReadSession struct {
	LastSeen time.Time `json:"lastSeen"`
}

type Invitation struct {
	Token   InvitationToken `json:"invitation"`
	Created time.Time       `json:"created"`
}

type TransactionProvider interface {
	Read(handler TransactionHandler) error
	Write(handler TransactionHandler) error
}

type TransactionHandler func(repositories *TransactableRepositories) error

type TransactableRepositories struct {
	Invitations InvitationRepository
	Users       UserRepository
}

type Auth struct {
	RegisterInitial  *RegisterInitialHandler
	Register         *RegisterHandler
	Login            *LoginHandler
	Logout           *LogoutHandler
	CheckAccessToken *CheckAccessTokenHandler
	List             *ListHandler
	CreateInvitation *CreateInvitationHandler
	Remove           *RemoveHandler
	SetPassword      *SetPasswordHandler
}

const maxUsernameLen = 100
const maxPasswordLen = 10000

func validate(username, password string) error {
	if username == "" {
		return errors.New("username can't be empty")
	}

	if len(username) > maxUsernameLen {
		return fmt.Errorf("username length can't exceed %d characters", maxUsernameLen)
	}

	if password == "" {
		return errors.New("password can't be empty")
	}

	if len(password) > maxPasswordLen {
		return fmt.Errorf("password length can't exceed %d characters", maxPasswordLen)
	}

	return nil
}

func toReadUsers(users []User) []ReadUser {
	var rv []ReadUser
	for _, user := range users {
		rv = append(rv, toReadUser(user))
	}
	return rv
}

func toReadUser(user User) ReadUser {
	rv := ReadUser{
		Username:      user.Username,
		Administrator: user.Administrator,
		Created:       user.Created,
		LastSeen:      user.LastSeen,
	}
	for _, session := range user.Sessions {
		rv.Sessions = append(rv.Sessions, ReadSession{
			LastSeen: session.LastSeen,
		})
	}
	return rv
}
