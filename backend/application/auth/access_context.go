package auth

import authdomain "github.com/boreq/eggplant/domain/auth"

type Permission struct {
	value string
}

func (p Permission) String() string {
	return p.value
}

var (
	PermissionManageUsers       = Permission{value: "manage_users"}
	PermissionCreateInvitations = Permission{value: "create_invitations"}
)

type AccessContext interface {
	Can(p Permission) bool
}

type AuthenticatedAccessContext struct {
	username      authdomain.Username
	administrator bool
	token         authdomain.AccessToken
}

func NewAuthenticatedAccessContext(u authdomain.User, t authdomain.AccessToken) AuthenticatedAccessContext {
	return AuthenticatedAccessContext{
		username:      u.Username(),
		administrator: u.Administrator(),
		token:         t,
	}
}

func (c AuthenticatedAccessContext) Can(Permission) bool {
	return c.administrator
}

func (c AuthenticatedAccessContext) Username() authdomain.Username {
	return c.username
}

func (c AuthenticatedAccessContext) Token() authdomain.AccessToken {
	return c.token
}

type CommandLineAccessContext struct{}

func NewCommandLineAccessContext() CommandLineAccessContext {
	return CommandLineAccessContext{}
}

func (CommandLineAccessContext) Can(Permission) bool {
	return true
}

type AnonymousAccessContext struct{}

func NewAnonymousAccessContext() AnonymousAccessContext {
	return AnonymousAccessContext{}
}

func (AnonymousAccessContext) Can(Permission) bool {
	return false
}
