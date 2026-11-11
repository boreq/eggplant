package auth

import (
	"errors"
	"time"

	authdomain "github.com/boreq/eggplant/domain/auth"
)

var (
	ErrUnauthorized         = errors.New("unauthorized")
	ErrUsernameTaken        = errors.New("username taken")
	ErrNotFound             = errors.New("not found")
	ErrAlreadyAuthenticated = errors.New("already authenticated")
	ErrCannotRemoveSelf     = errors.New("cannot remove self")
	ErrUsersAlreadyExist    = errors.New("there are existing users")
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

	// GetByToken returns the user owning the session with the provided
	// access token. If no such session exists ErrNotFound is returned.
	GetByToken(token authdomain.AccessToken) (*authdomain.User, error)

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

type LastSeenUpdate struct {
	username authdomain.Username
	lastSeen time.Time
	sessions map[authdomain.AccessToken]time.Time
}

func NewLastSeenUpdate(username authdomain.Username, lastSeen time.Time, sessions map[authdomain.AccessToken]time.Time) LastSeenUpdate {
	return LastSeenUpdate{
		username: username,
		lastSeen: lastSeen,
		sessions: sessions,
	}
}

func (u LastSeenUpdate) Username() authdomain.Username {
	return u.username
}

func (u LastSeenUpdate) LastSeen() time.Time {
	return u.lastSeen
}

func (u LastSeenUpdate) Sessions() map[authdomain.AccessToken]time.Time {
	return u.sessions
}

type LastSeenUpdater interface {
	Update(username authdomain.Username, token authdomain.AccessToken, t time.Time)
	PopUpdates() []LastSeenUpdate
}

type TransactionProvider interface {
	Read(handler TransactionHandler) error
	Write(handler TransactionHandler) error
}

type TransactionHandler func(repositories *TransactableRepositories) error

type TransactableRepositories struct {
	invitations InvitationRepository
	users       UserRepository
}

func NewTransactableRepositories(
	invitations InvitationRepository,
	users UserRepository,
) *TransactableRepositories {
	return &TransactableRepositories{
		invitations: invitations,
		users:       users,
	}
}

func (r TransactableRepositories) Invitations() InvitationRepository {
	return r.invitations
}

func (r TransactableRepositories) Users() UserRepository {
	return r.users
}

type Auth struct {
	registerInitial  *RegisterInitialHandler
	register         *RegisterHandler
	login            *LoginHandler
	logout           *LogoutHandler
	checkAccessToken *CheckAccessTokenHandler
	list             *ListHandler
	createInvitation *CreateInvitationHandler
	remove           *RemoveHandler
	setPassword      *SetPasswordHandler
	getCurrentUser   *GetCurrentUserHandler
}

func NewAuth(
	registerInitial *RegisterInitialHandler,
	register *RegisterHandler,
	login *LoginHandler,
	logout *LogoutHandler,
	checkAccessToken *CheckAccessTokenHandler,
	list *ListHandler,
	createInvitation *CreateInvitationHandler,
	remove *RemoveHandler,
	setPassword *SetPasswordHandler,
	getCurrentUser *GetCurrentUserHandler,
) *Auth {
	return &Auth{
		registerInitial:  registerInitial,
		register:         register,
		login:            login,
		logout:           logout,
		checkAccessToken: checkAccessToken,
		list:             list,
		createInvitation: createInvitation,
		remove:           remove,
		setPassword:      setPassword,
		getCurrentUser:   getCurrentUser,
	}
}

func (a Auth) RegisterInitial() *RegisterInitialHandler {
	return a.registerInitial
}

func (a Auth) Register() *RegisterHandler {
	return a.register
}

func (a Auth) Login() *LoginHandler {
	return a.login
}

func (a Auth) Logout() *LogoutHandler {
	return a.logout
}

func (a Auth) CheckAccessToken() *CheckAccessTokenHandler {
	return a.checkAccessToken
}

func (a Auth) List() *ListHandler {
	return a.list
}

func (a Auth) CreateInvitation() *CreateInvitationHandler {
	return a.createInvitation
}

func (a Auth) Remove() *RemoveHandler {
	return a.remove
}

func (a Auth) SetPassword() *SetPasswordHandler {
	return a.setPassword
}

func (a Auth) GetCurrentUser() *GetCurrentUserHandler {
	return a.getCurrentUser
}
