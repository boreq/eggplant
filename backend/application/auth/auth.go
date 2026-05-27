package auth

import (
	"errors"
	"time"

	authdomain "github.com/boreq/eggplant/domain/auth"
)

var (
	ErrUnauthorized  = errors.New("unauthorized")
	ErrUsernameTaken = errors.New("username taken")
	ErrNotFound      = errors.New("not found")
)

type PasswordHasher interface {
	Hash(password authdomain.Password) (authdomain.PasswordHash, error)
	Compare(hashedPassword authdomain.PasswordHash, password authdomain.Password) error
}

type UserRepository interface {
	// Put inserts the user into the repository. The previous entry with
	// this username is overwriten.
	Put(user authdomain.User) error

	// Get returns the user with the provided username. If the user doesn't
	// exist ErrNotFound is returned.
	Get(username authdomain.Username) (*authdomain.User, error)

	// Remove should remove a user.
	Remove(username authdomain.Username) error

	// List should return a list of all users.
	List() ([]authdomain.User, error)

	// Count returns the number of users.
	Count() (int, error)
}

type InvitationRepository interface {
	// Put inserts the invitation into the repository. The previous entry
	// with this token is overwriten.
	Put(invitation authdomain.Invitation) error

	// Get returns an invitation with the provided token, if the invitation
	// doesn't exist ErrNotFound is returned.
	Get(token authdomain.InvitationToken) (*authdomain.Invitation, error)

	// Remove removes an invitation. If the invitation doesn't exist this
	// function returns nil.
	Remove(token authdomain.InvitationToken) error
}

type LastSeenUpdater interface {
	Update(username authdomain.Username, token authdomain.AccessToken, t time.Time)
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
